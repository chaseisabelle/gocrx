package gocrx

import (
	"context"
	"fmt"
	"github.com/mediabuyerbot/go-crx3"
	"os"
)

type File struct {
	logger    Logger
	extension crx3.Extension
	directory *Directory
}

func NewFile(pth string, lgr Logger) (*File, error) {
	if lgr == nil {
		lgr = func(context.Context, error) {}
	}

	tmp, err := os.MkdirTemp("", "gocrx-*")

	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	ext := crx3.Extension(pth)

	switch {
	case ext.IsZip():
		err = crx3.UnzipTo(tmp, pth)
	case ext.IsCRX3():
		err = crx3.UnpackTo(pth, tmp)
	default:
		err = fmt.Errorf("invalid extension %s", pth)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to extract extension: %w", err)
	}

	dir, err := NewDirectory(tmp, lgr)

	if err != nil {
		return nil, fmt.Errorf("failed to create directory extension: %w", err)
	}

	return &File{
		logger:    lgr,
		extension: ext,
		directory: dir,
	}, nil
}

func (f *File) IsZip() bool {
	return f.extension.IsZip()
}

func (f *File) IsCRX3() bool {
	return f.extension.IsCRX3()
}

func (f *File) Manifest(ctx context.Context) (Manifest, error) {
	return f.directory.Manifest(ctx)
}

func (f *File) Close(ctx context.Context) error {
	err := f.directory.Close(ctx)

	if err != nil {
		return fmt.Errorf("failed to close directory extension: %w", err)
	}

	err = os.RemoveAll(f.directory.extension.String())

	if err != nil {
		return fmt.Errorf("failed to remove temporary directory: %w", err)
	}

	return nil
}
