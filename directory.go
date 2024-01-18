package gocrx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mediabuyerbot/go-crx3"
	"os"
	"path/filepath"
)

type Directory struct {
	logger    Logger
	extension crx3.Extension
}

func NewDirectory(dir string, lgr Logger) (*Directory, error) {
	if lgr == nil {
		lgr = func(context.Context, error) {}
	}

	ext := crx3.Extension(dir)

	if !ext.IsDir() {
		return nil, fmt.Errorf("invalid directory %s", dir)
	}

	return &Directory{
		logger:    lgr,
		extension: ext,
	}, nil
}

func (d *Directory) Manifest(context.Context) (Manifest, error) {
	buf, err := os.ReadFile(filepath.Join(d.extension.String(), "manifest.json"))

	if err != nil {
		return Manifest{}, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var man Manifest

	err = json.Unmarshal(buf, &man)

	if err != nil {
		return man, fmt.Errorf("failed to parse manifest file: %w", err)
	}

	return man, nil
}

func (d *Directory) Close(context.Context) error {
	return nil
}
