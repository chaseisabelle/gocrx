package gocrx

import (
	"context"
	"fmt"
	"github.com/zserge/lorca"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

type Chrome struct {
	logger Logger
}

func NewChrome(lgr Logger) *Chrome {
	if lgr == nil {
		lgr = func(context.Context, error) {}
	}

	return &Chrome{
		logger: lgr,
	}
}

func (c *Chrome) PackAndSign(ctx context.Context, ext Extension) ([]byte, string, error) {
	return nil, "", nil
}

func (c *Chrome) PackWithPEM(ctx context.Context, ext Extension, pem string) ([]byte, error) {
	return nil, nil
}

func (c *Chrome) pack(ctx context.Context, dir string, pem string) ([]byte, string, error) {
	bin := lorca.LocateChrome()

	if bin == "" {
		return nil, "", fmt.Errorf("failed to run chromium/chrome: failed to locate binary")
	}

	arg := []string{
		"--no-sandbox",
		fmt.Sprintf("--pack-extension=%s", dir),
	}

	if pem != "" {
		arg = append(arg, fmt.Sprintf("--pack-extension-key=%s", pem))
	} else {
		pem = filepath.Join(dir, "ext.pem")
	}

	cmd := exec.Command(bin, arg...)

	sep, err := cmd.StderrPipe()

	if err != nil {
		c.logger(ctx, fmt.Errorf("failed to grab chromium/chrome stderr: %w", err))
	}

	sop, err := cmd.StdoutPipe()

	if err != nil {
		c.logger(ctx, fmt.Errorf("failed to grab chromium/chrome stdout: %w", err))
	}

	err = cmd.Start()

	if err != nil {
		return nil, "", fmt.Errorf("failed to run chromium/chrome: %w", err)
	}

	var lll string

	if sep != nil {
		buf, err := io.ReadAll(sep)

		if err != nil {
			c.logger(ctx, fmt.Errorf("failed to read chromium/chrome stderr: %w", err))
		} else {
			lll = strings.TrimSpace(string(buf))
		}
	}

	if sop != nil {
		buf, err := io.ReadAll(sop)

		if err != nil {
			c.logger(ctx, fmt.Errorf("failed to read chromium/chrome stdout: %w", err))
		} else {
			lll += "\n" + strings.TrimSpace(string(buf))
		}
	}

	if lll == "" {
		lll = "no stderr/stdout"
	}

	err = cmd.Wait()

	if err != nil {
		return nil, "", fmt.Errorf("failed to run chromium: %w: %s", err, lll)
	}

	err = renameFile(crx, filepath.Join(dir, "ext.crx"))

	if err != nil {
		return out, fmt.Errorf("failed to rename crx file: %w", err)
	}
}
