package gocrx

import (
	"context"
	"encoding/json"
	"fmt"
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

// PackInput is the input to the Pack function.
type PackInput struct {
	// Packer is the type of packer to use.
	Packer Packer
	// Directory is the directory of the extension.
	Directory string
	// ZipFile is the zip file of the extension.
	ZipFile string
	// ZipBytes is the zip bytes of the extension.
	ZipBytes []byte
	// PEMFile is the pem file of the extension.
	PEMFile string
	// PEMBytes is the pem bytes of the extension.
	PEMBytes []byte
}

// PackOptions is the options for the Pack function.
type PackOptions struct {
	// OnError is the non-critical errors handler.
	OnError func(context.Context, error)
	// Sign is to sign/re-sign the crx (generates new key+pem).
	Sign bool
	// Version is the version to set in the manifest.
	Version string
}

// PackOutput is the output of the Pack function.
type PackOutput struct {
	// Manifest is the manifest data of the packed extension.
	Manifest map[string]any
	// PEMBytes is the pem bytes of the packed extension.
	PEMBytes []byte
	// CRXBytes is the crx bytes of the packed extension.
	CRXBytes []byte
}

const (
	// Chrome uses and requires chromium/chrome to be installed.
	Chrome Packer = "chrome"
	// GoCRX3 is a purely go implementation of the crx3 format.
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

	if len(inp.ZipBytes) > 0 {
		if inp.ZipFile != "" {
			return out, fmt.Errorf("cannot use both zip file and zip bytes")
		}

		inp.ZipFile = filepath.Join(dir, "ext.zip")

		err = writeBytesToFile(ctx, opt.OnError, inp.ZipFile, inp.ZipBytes)

		if err != nil {
			return out, fmt.Errorf("failed to write zip bytes to file: %w", err)
		}
	}

	if inp.ZipFile != "" {
		if inp.Directory != "" {
			return out, fmt.Errorf("cannot use both directory and zip file/bytes")
		}

		inp.Directory = filepath.Join(dir, "ext")

		err = crx3.Extension(inp.ZipFile).Unzip()

		if err != nil {
			return out, fmt.Errorf("failed to unzip to dir: %w", err)
		}

		err = renameFile(ctx, opt.OnError, inp.Directory[:len(inp.Directory)-4], dir)

		if err != nil {
			return out, fmt.Errorf("failed to rename dir: %w", err)
		}
	}

	if inp.Directory == "" {
		return out, fmt.Errorf("input dir, zip file, or zip bytes required")
	}

	ext := filepath.Join(dir, "ext")
	err = copier.Copy(inp.Directory, ext)

	if err != nil {
		return out, fmt.Errorf("failed to copy dir: %w", err)
	}

	inp.Directory = ext

	if len(inp.PEMBytes) > 0 {
		if inp.PEMFile != "" {
			return out, fmt.Errorf("cannot use both pem file and pem bytes")
		}

		inp.PEMFile = filepath.Join(dir, "ext.pem")

		err = writeBytesToFile(ctx, opt.OnError, inp.PEMFile, inp.PEMBytes)

		if err != nil {
			return out, fmt.Errorf("failed to write pem bytes to file: %w", err)
		}
	}

	if inp.PEMFile == "" && !opt.Sign {
		return out, fmt.Errorf("input pem file or pem bytes required or sign must be true")
	}

	if inp.PEMFile != "" && opt.Sign {
		return out, fmt.Errorf("cannot use both pem file and sign")
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

	if opt.Version != "" {
		out.Manifest["version"] = opt.Version
	}

	buf, err = json.MarshalIndent(out.Manifest, "", "  ")

	if err != nil {
		return out, fmt.Errorf("failed to marshal manifest data: %w", err)
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
			arg = append(arg, fmt.Sprintf("--pack-extension-key=%s", inp.PEMFile))
		} else {
			inp.PEMFile = filepath.Join(dir, "ext.pem")
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

		err = renameFile(ctx, opt.OnError, crx, filepath.Join(dir, "ext.crx"))

		if err != nil {
			return out, fmt.Errorf("failed to rename crx file: %w", err)
		}
	case GoCRX3:
		if opt.Sign {
			return out, fmt.Errorf("signing not supported with gocrx3...yet")
		}

		key, err := crx3.LoadPrivateKey(inp.PEMFile)

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

	out.CRXBytes, err = os.ReadFile(crx)

	if err != nil {
		return out, fmt.Errorf("failed to read crx file: %w", err)
	}

	out.PEMBytes, err = os.ReadFile(inp.PEMFile)

	if err != nil {
		return out, fmt.Errorf("failed to read pem file: %w", err)
	}

	return out, nil
}
