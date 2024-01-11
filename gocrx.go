package gocrx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func renameFile(ctx context.Context, lgr func(context.Context, error), old string, new string) error {
	if old == new {
		return nil
	}

	err := os.Rename(old, new)

	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

func writeBytesToFile(ctx context.Context, lgr func(context.Context, error), pth string, buf []byte) error {
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
