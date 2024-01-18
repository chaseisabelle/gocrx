package gocrx

import (
	"context"
	"fmt"
	"os"
)

type Binary struct {
	logger Logger
	file   *File
}

func NewBinary(bin []byte, lgr Logger) (*Binary, error) {
	if lgr == nil {
		lgr = func(context.Context, error) {}
	}

	tmp, err := os.CreateTemp("", "gocrx-*")

	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	defer func() {
		nam := tmp.Name()
		err := tmp.Close()

		if err != nil {
			lgr(context.Background(), fmt.Errorf("failed to close temporary file: %w", err))
		}

		err = os.RemoveAll(nam)

		if err != nil {
			lgr(context.Background(), fmt.Errorf("failed to remove temporary file: %w", err))
		}
	}()

	inf, err := tmp.Stat()

	if err != nil {
		return nil, fmt.Errorf("failed to stat temporary file: %w", err)
	}

	err = os.WriteFile(tmp.Name(), bin, inf.Mode().Perm())

	if err != nil {
		return nil, fmt.Errorf("failed to write temporary file: %w", err)
	}

	fil, err := NewFile(tmp.Name(), lgr)

	if err != nil {
		return nil, fmt.Errorf("failed to create file extension: %w", err)
	}

	return &Binary{
		file: fil,
	}, nil
}

func (b *Binary) Manifest(ctx context.Context) (Manifest, error) {
	return b.file.Manifest(ctx)
}

func (b *Binary) IsZip() bool {
	return b.file.IsZip()
}

func (b *Binary) IsCRX3() bool {
	return b.file.IsCRX3()
}

func (b *Binary) Close(ctx context.Context) error {
	err := b.file.Close(ctx)

	if err != nil {
		return fmt.Errorf("failed to close file extension: %w", err)
	}

	return nil
}
