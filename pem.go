package gocrx

import (
	"errors"
	"fmt"
	"os"
)

type PEM struct {
	internal string
}

func NewPEM(pem string) (*PEM, error) {
	_, err := os.Stat(pem)

	if errors.Is(err, os.ErrNotExist) {
		return &PEM{
			internal: pem,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to stat pem file: %w", err)
	}

	buf, err := os.ReadFile(pem)

	if err != nil {
		return nil, fmt.Errorf("failed to read pem file: %w", err)
	}

	return &PEM{
		internal: string(buf),
	}, nil
}
