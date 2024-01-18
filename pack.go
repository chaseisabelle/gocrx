package gocrx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mediabuyerbot/go-crx3"
	copier "github.com/otiai10/copy"
	"github.com/zserge/lorca"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Packer is the type of packer to use.
type Packer string

// PackInput is the input to the Pack function.
type PackInput struct {
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

// PackOptions is the options for the Pack function.
type PackOptions struct {
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

// PackOutput is the output of the Pack function.
type PackOutput struct {
	// Manifest is the manifest data of the packed extension.
	Manifest map[string]any
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
func Pack(ctx context.Context, inp PackInput, opt PackOptions) (PackOutput, error) {
	var out PackOutput

	if inp.Packer == "" {
		inp.Packer = Chrome
	}

	if opt.OnError == nil {
		opt.OnError = func(context.Context, error) {}
	}

	dir, err := os.MkdirTemp("", "gocrx-*")

	if err != nil {
		return out, fmt.Errorf("failed to create temp dir: %w", err)
	}

	defer func() {
		err := os.RemoveAll(dir)

		if err != nil {
			opt.OnError(ctx, fmt.Errorf("failed to remove workdir: %w", err))
		}
	}()

	if len(inp.Binary) > 0 {
		if inp.File != "" {
			return out, fmt.Errorf("cannot use both file and binary")
		}

		inp.File = filepath.Join(dir, "bin")

		err = writeBytesToFile(inp.File, inp.Binary)

		if err != nil {
			return out, fmt.Errorf("failed to write binary to file: %w", err)
		}
	}

	if inp.File != "" {
		if inp.Directory != "" {
			return out, fmt.Errorf("cannot use both directory and file/binary")
		}

		inp.Directory = filepath.Join(dir, "ext")

		ext := crx3.Extension(inp.File)

		switch {
		case ext.IsZip():
			err = ext.Unzip()
		case ext.IsCRX3():
			err = ext.Unpack()
		default:
			err = fmt.Errorf("invalid file/binary")
		}

		if err != nil {
			return out, fmt.Errorf("failed to extract file/binary to dir: %w", err)
		}

		err = renameFile(inp.Directory[:len(inp.Directory)-4], dir)

		if err != nil {
			return out, fmt.Errorf("failed to rename dir: %w", err)
		}
	}

	if inp.Directory == "" {
		return out, fmt.Errorf("input dir/file/binary required")
	}

	ext := filepath.Join(dir, "ext")
	err = copier.Copy(inp.Directory, ext)

	if err != nil {
		return out, fmt.Errorf("failed to copy dir: %w", err)
	}

	inp.Directory = ext

	if inp.PEM != "" {
		_, err = os.Stat(inp.PEM)

		if errors.Is(err, fs.ErrNotExist) {
			inp.PEM = filepath.Join(dir, "ext.pem")

			err = writeBytesToFile(filepath.Join(dir, inp.PEM), []byte(inp.PEM))
		}

		if err != nil {
			return out, fmt.Errorf("failed to utilize pem input: %w", err)
		}
	}

	if inp.PEM == "" && !opt.Sign {
		return out, fmt.Errorf("input pem required or sign must be true")
	}

	if inp.PEM != "" && opt.Sign {
		return out, fmt.Errorf("cannot use both pem and sign option")
	}

	crx := filepath.Join(dir, "ext.crx")

	buf, err := os.ReadFile(filepath.Join(inp.Directory, "manifest.json"))

	if err != nil {
		return out, fmt.Errorf("failed to read manifest file: %w", err)
	}

	err = json.Unmarshal(buf, &out.Manifest)

	if err != nil {
		return out, fmt.Errorf("failed to unmarshal manifest data: %w", err)
	}

	if opt.Sign {
		delete(out.Manifest, "key")
	}

	if opt.Name != "" {
		out.Manifest["name"] = opt.Name
	}

	if opt.ShortName != "" {
		out.Manifest["short_name"] = opt.ShortName
	}

	if opt.Description != "" {
		out.Manifest["description"] = opt.Description
	}

	if opt.Version != "" {
		out.Manifest["version"] = opt.Version
	}

	buf, err = json.MarshalIndent(out.Manifest, "", "  ")

	if err != nil {
		return out, fmt.Errorf("failed to marshal manifest data: %w", err)
	}

	err = writeBytesToFile(filepath.Join(inp.Directory, "manifest.json"), buf)

	if err != nil {
		return out, fmt.Errorf("failed to write manifest file: %w", err)
	}

	switch inp.Packer {
	case Chrome:
		bin := lorca.LocateChrome()

		if bin == "" {
			return out, fmt.Errorf("failed to run chromium/chrome: failed to locate binary")
		}

		arg := []string{
			"--no-sandbox",
			fmt.Sprintf("--pack-extension=%s", inp.Directory),
		}

		if !opt.Sign {
			arg = append(arg, fmt.Sprintf("--pack-extension-key=%s", inp.PEM))
		} else {
			inp.PEM = filepath.Join(dir, "ext.pem")
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
			return out, fmt.Errorf("failed to run chromium/chrome: %w", err)
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
			return out, fmt.Errorf("failed to run chromium: %w: %s", err, lll)
		}

		err = renameFile(crx, filepath.Join(dir, "ext.crx"))

		if err != nil {
			return out, fmt.Errorf("failed to rename crx file: %w", err)
		}
	case GoCRX3:
		if opt.Sign {
			npk, err := crx3.NewPrivateKey()

			if err != nil {
				return out, fmt.Errorf("failed to generate private key: %w", err)
			}

			inp.PEM = filepath.Join(dir, "ext.pem")

			err = crx3.SavePrivateKey(inp.PEM, npk)

			if err != nil {
				return out, fmt.Errorf("failed to save private key: %w", err)
			}
		}

		key, err := crx3.LoadPrivateKey(inp.PEM)

		if err != nil {
			return out, fmt.Errorf("failed to load private key: %w", err)
		}

		err = crx3.Extension(inp.Directory).PackTo(crx, key)

		if err != nil {
			return out, fmt.Errorf("failed to pack extension: %w", err)
		}
	default:
		return out, fmt.Errorf("invalid packer '%s'", inp.Packer)
	}

	out.CRX, err = os.ReadFile(crx)

	if err != nil {
		return out, fmt.Errorf("failed to read crx file: %w", err)
	}

	buf, err = os.ReadFile(inp.PEM)

	if err != nil {
		return out, fmt.Errorf("failed to read pem file: %w", err)
	}

	out.PEM = string(buf)

	return out, nil
}

func renameFile(old string, new string) error {
	if old == new {
		return nil
	}

	err := os.Rename(old, new)

	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
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
