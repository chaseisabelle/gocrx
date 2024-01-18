package gocrx

import (
	"context"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/mediabuyerbot/go-crx3"
	copier "github.com/otiai10/copy"
	"github.com/zserge/lorca"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Packer is the type of packer to use.
type Packer string

// Input is the input to the Pack function.
type Input struct {
	// Packer is the type of packer to use.
	Packer Packer
	// Directory is the directory of the extension.
	Directory string
	// File is the zip/crx file of the extension.
	File string
	// Binary is the zip/crx bytes of the extension.
	Binary []byte
	// PEM is the pem key/file of the extension.
	PEM string
}

// Options is the options for the Pack function.
type Options struct {
	// OnError is the non-critical errors handler.
	OnError func(context.Context, error)
	// Sign is to sign/re-sign the crx (generates new key+pem).
	Sign bool
	// Name is the name to set in the manifest.
	Name string
	// ShortName is the short name to set in the manifest.
	ShortName string
	// Description is the description to set in the manifest.
	Description string
	// Version is the version to set in the manifest.
	Version string
}

// Output is the output of the Pack function.
type Output struct {
	// Manifest is the manifest data of the packed extension.
	Manifest *simplejson.Json
	// PEM is the pem key of the packed extension.
	PEM string
	// CRX is the crx binary of the packed extension.
	CRX []byte
}

const (
	// Chrome uses and requires chromium/chrome to be installed.
	Chrome Packer = "chrome"
	// GoCRX3 uses pure go: https://github.com/mmadfox/go-crx3
	GoCRX3 Packer = "gocrx3"
)

// Pack packs an extension.
func Pack(ctx context.Context, inp Input, opt Options) (Output, error) {
	// initialize some required variables
	var out Output

	if inp.Packer == "" {
		inp.Packer = Chrome
	}

	if opt.OnError == nil {
		opt.OnError = func(context.Context, error) {}
	}

	// create temporary working directory
	dir, err := os.MkdirTemp("", "gocrx-*")

	if err != nil {
		return out, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// delete temporary working directory and its contents
	defer func() {
		err := os.RemoveAll(dir)

		if err != nil {
			opt.OnError(ctx, fmt.Errorf("failed to remove temporary directory: %w", err))
		}
	}()

	// create temporary working directory paths
	tbf := filepath.Join(dir, "bin")
	txd := filepath.Join(dir, "dir")
	tmf := filepath.Join(txd, "manifest.json")
	tpf := filepath.Join(dir, "pem")
	tcf := filepath.Join(dir, "crx")

	// check if the user has supplied a binary input
	if len(inp.Binary) > 0 {
		if inp.File != "" {
			return out, fmt.Errorf("cannot supply both file and binary inputs")
		}

		inp.File = tbf

		err = writeBytesToFile(inp.File, inp.Binary)

		if err != nil {
			return out, fmt.Errorf("failed to write binary to temporary file: %w", err)
		}
	}

	// check if the user has supplied a file input
	if inp.File != "" {
		if inp.Directory != "" {
			return out, fmt.Errorf("cannot supply both directory and file/binary inputs")
		}

		inp.Directory = txd

		tmp := crx3.Extension(inp.File)

		switch {
		case tmp.IsZip():
			err = os.Mkdir(inp.Directory, 0755)

			if err != nil {
				return out, fmt.Errorf("failed to create directory: %w", err)
			}

			err = crx3.UnzipTo(inp.Directory, inp.File)
		case tmp.IsCRX3():
			err = crx3.UnpackTo(inp.File, inp.Directory)
		default:
			err = fmt.Errorf("invalid file/binary")
		}

		if err != nil {
			return out, fmt.Errorf("failed to extract file/binary to directory: %w", err)
		}

		inp.Directory = filepath.Join(inp.Directory, filepath.Base(inp.File))
	}

	// check if the user has supplied a directory input
	if inp.Directory == "" {
		return out, fmt.Errorf("failed to supply directory/file/binary input")
	}

	// if the directory input is not the temporary directory, copy it to the temporary directory
	if inp.Directory != txd {
		err = copier.Copy(inp.Directory, txd)

		if err != nil {
			return out, fmt.Errorf("failed to copy directory: %w", err)
		}

		inp.Directory = txd
	}

	// check if the user has supplied a pem input
	if inp.PEM == "" {
		if !opt.Sign {
			return out, fmt.Errorf("pem input or sign option required")
		}

		npk, err := crx3.NewPrivateKey()

		if err != nil {
			return out, fmt.Errorf("failed to generate private key: %w", err)
		}

		inp.PEM = tpf

		err = crx3.SavePrivateKey(inp.PEM, npk)

		if err != nil {
			return out, fmt.Errorf("failed to save private key: %w", err)
		}
	}

	_, err = os.Stat(inp.PEM)

	if err != nil {
		err = writeBytesToFile(tpf, []byte(inp.PEM))

		if err != nil {
			return out, fmt.Errorf("failed to write temporary pem file: %w", err)
		}

		inp.PEM = tpf
	}

	// read+parse the manifest file and apply the options
	buf, err := os.ReadFile(tmf)

	if err != nil {
		return out, fmt.Errorf("failed to read manifest file: %w", err)
	}

	out.Manifest, err = simplejson.NewJson(buf)

	if err != nil {
		return out, fmt.Errorf("failed to parse manifest data: %w", err)
	}

	if opt.Sign {
		out.Manifest.Del("key")
	}

	if opt.Name != "" {
		out.Manifest.Set("name", opt.Name)
	}

	if opt.ShortName != "" {
		out.Manifest.Set("short_name", opt.ShortName)
	}

	if opt.Description != "" {
		out.Manifest.Set("description", opt.Description)
	}

	if opt.Version != "" {
		out.Manifest.Set("version", opt.Version)
	}

	// encode+write the manifest file
	buf, err = out.Manifest.EncodePretty()

	if err != nil {
		return out, fmt.Errorf("failed to encode manifest data: %w", err)
	}

	err = writeBytesToFile(tmf, buf)

	if err != nil {
		return out, fmt.Errorf("failed to write manifest file: %w", err)
	}

	// pack the extension using the specified packer
	switch inp.Packer {
	case Chrome:
		bin := lorca.LocateChrome()

		if bin == "" {
			return out, fmt.Errorf("failed to locate chrome/chromium binary")
		}

		arg := []string{
			"--no-sandbox",
			fmt.Sprintf("--pack-extension=%s", inp.Directory),
		}

		if !opt.Sign {
			arg = append(arg, fmt.Sprintf("--pack-extension-key=%s", inp.PEM))
		} else {
			inp.PEM = tpf
		}

		cmd := exec.Command(bin, arg...)

		sep, err := cmd.StderrPipe()

		if err != nil {
			opt.OnError(ctx, fmt.Errorf("failed to grab chromium/chrome stderr: %w", err))
		}

		sop, err := cmd.StdoutPipe()

		if err != nil {
			opt.OnError(ctx, fmt.Errorf("failed to grab chromium/chrome stdout: %w", err))
		}

		err = cmd.Start()

		if err != nil {
			return out, fmt.Errorf("failed to run chromium/chrome '%s': %w", bin, err)
		}

		var lll string

		if sep != nil {
			buf, err := io.ReadAll(sep)

			if err != nil {
				opt.OnError(ctx, fmt.Errorf("failed to read chromium/chrome stderr: %w", err))
			} else {
				lll = strings.TrimSpace(string(buf))
			}
		}

		if sop != nil {
			buf, err := io.ReadAll(sop)

			if err != nil {
				opt.OnError(ctx, fmt.Errorf("failed to read chromium/chrome stdout: %w", err))
			} else {
				lll += "\n" + strings.TrimSpace(string(buf))
			}
		}

		if lll == "" {
			lll = "no stderr/stdout"
		}

		err = cmd.Wait()

		if err != nil {
			return out, fmt.Errorf("failed to run chrome/chromium '%s': %w: %s", bin, err, lll)
		}

		tcf = filepath.Join(dir, fmt.Sprintf("%s.crx", filepath.Base(txd)))
	case GoCRX3:
		key, err := crx3.LoadPrivateKey(inp.PEM)

		if err != nil {
			return out, fmt.Errorf("failed to load private key: %w", err)
		}

		tcf = fmt.Sprintf("%s.crx", tcf)
		err = crx3.Extension(inp.Directory).PackTo(tcf, key)

		if err != nil {
			return out, fmt.Errorf("failed to pack extension: %w", err)
		}
	default:
		return out, fmt.Errorf("invalid packer '%s'", inp.Packer)
	}

	// read the crx file and pem file
	out.CRX, err = os.ReadFile(tcf)

	if err != nil {
		return out, fmt.Errorf("failed to read crx file: %w", err)
	}

	buf, err = os.ReadFile(inp.PEM)

	if err != nil {
		return out, fmt.Errorf("failed to read pem file: %w", err)
	}

	out.PEM = string(buf)

	// kthxbye
	return out, nil
}

func writeBytesToFile(pth string, buf []byte) error {
	inf, err := os.Stat(pth)

	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to write file: %w", err)
		}

		inf, err = os.Stat(filepath.Dir(pth))

		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	err = os.WriteFile(pth, buf, inf.Mode().Perm())

	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
