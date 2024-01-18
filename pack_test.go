package gocrx_test

import (
	"context"
	"github.com/chaseisabelle/gocrx"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func packOptions(t *testing.T, opt gocrx.PackOptions) gocrx.PackOptions {
	opt.OnError = func(_ context.Context, err error) {
		assert.NoError(t, err)
	}

	return opt
}

func TestPack_ChromePacker_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	inp := gocrx.PackInput{
		Packer:    gocrx.Chrome,
		Directory: "tmp/test",
	}

	opt := packOptions(t, gocrx.PackOptions{
		Sign: true,
	})

	tc1, err := gocrx.Pack(ctx, inp, opt)

	if !assert.NoError(t, err) || !assert.NotNil(t, tc1.Manifest) || !assert.NotEmpty(t, tc1.PEM) || !assert.NotEmpty(t, tc1.CRX) {
		return
	}

	err = os.WriteFile("tmp/test.crx", tc1.CRX, 0644)

	if !assert.NoError(t, err) {
		return
	}

	err = os.WriteFile("tmp/test.pem", []byte(tc1.PEM), 0644)

	if !assert.NoError(t, err) {
		return
	}

	inp = gocrx.PackInput{
		Packer:    gocrx.Chrome,
		Directory: "tmp/test",
		PEM:       "tmp/test.pem",
	}

	opt = packOptions(t, gocrx.PackOptions{})

	tc2, err := gocrx.Pack(ctx, inp, opt)

	if !assert.NoError(t, err) || !assert.Equal(t, tc1.Manifest, tc2.Manifest) || !assert.Equal(t, tc1.PEM, tc2.PEM) || !assert.Equal(t, tc1.CRX, tc2.CRX) {
		return
	}

	inp = gocrx.PackInput{
		Packer: gocrx.Chrome,
		Binary: tc1.CRX,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, gocrx.PackOptions{})

	tc3, err := gocrx.Pack(ctx, inp, opt)

	if !assert.NoError(t, err) || !assert.Equal(t, tc1.Manifest, tc3.Manifest) || !assert.Equal(t, tc1.PEM, tc3.PEM) || !assert.Equal(t, tc1.CRX, tc3.CRX) {
		return
	}
}

func TestPack_GoCRX3Packer_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	inp := gocrx.PackInput{
		Packer:    gocrx.GoCRX3,
		Directory: "tmp/test",
	}

	opt := packOptions(t, gocrx.PackOptions{
		Sign: true,
	})

	tc1, err := gocrx.Pack(ctx, inp, opt)

	if !assert.NoError(t, err) || !assert.NotNil(t, tc1.Manifest) || !assert.NotEmpty(t, tc1.PEM) || !assert.NotEmpty(t, tc1.CRX) {
		return
	}

	err = os.WriteFile("tmp/test.crx", tc1.CRX, 0644)

	if !assert.NoError(t, err) {
		return
	}

	err = os.WriteFile("tmp/test.pem", []byte(tc1.PEM), 0644)

	if !assert.NoError(t, err) {
		return
	}

	inp = gocrx.PackInput{
		Packer:    gocrx.GoCRX3,
		Directory: "tmp/test",
		PEM:       "tmp/test.pem",
	}

	opt = packOptions(t, gocrx.PackOptions{})

	tc2, err := gocrx.Pack(ctx, inp, opt)

	if !assert.NoError(t, err) || !assert.Equal(t, tc1.Manifest, tc2.Manifest) || !assert.Equal(t, tc1.PEM, tc2.PEM) || !assert.Equal(t, tc1.CRX, tc2.CRX) {
		return
	}
}
