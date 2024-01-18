package gocrx_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chaseisabelle/gocrx"
	"github.com/mediabuyerbot/go-crx3"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestPack_ChromePacker_Success(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "gocrx-test-*")

	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()

	ctx := context.Background()

	inp := &gocrx.Input{
		Packer:    gocrx.Chrome,
		Directory: "tmp/test",
	}

	opt := packOptions(t, &gocrx.Options{
		Sign: true,
	})

	tc1, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.NotNil(t, tc1.Manifest):
		fallthrough
	case !assert.NotEmpty(t, tc1.PEM):
		fallthrough
	case !assert.NotEmpty(t, tc1.CRX):
		return
	}

	tcf := filepath.Join(dir, "crx")

	if !assert.NoError(t, writeBytesToFile(tcf, tc1.CRX)) {
		return
	}

	tpf := filepath.Join(dir, "pem")

	if !assert.NoError(t, writeBytesToFile(tpf, []byte(tc1.PEM))) {
		return
	}

	inp = &gocrx.Input{
		Packer:    gocrx.Chrome,
		Directory: "tmp/test",
		PEM:       tpf,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc2, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc2.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc2.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc2.CRX)):
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.Chrome,
		Binary: tc1.CRX,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc3, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc3.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc3.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc3.CRX)):
		return
	}

	tzf := filepath.Join(dir, "zip")

	if !assert.NoError(t, crx3.ZipTo(tzf, "tmp/test")) {
		return
	}

	bin, err := os.ReadFile(tzf)

	if !assert.NoError(t, err) {
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.Chrome,
		Binary: bin,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc4, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc4.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc4.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc4.CRX)):
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.Chrome,
		File:   tcf,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc5, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc5.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc5.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc5.CRX)):
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.Chrome,
		File:   tzf,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc6, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc6.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc6.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc6.CRX)):
		return
	}
}

func TestPack_GoCRX3Packer_Success(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "gocrx-test-*")

	if !assert.NoError(t, err) {
		return
	}

	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()

	ctx := context.Background()

	inp := &gocrx.Input{
		Packer:    gocrx.GoCRX3,
		Directory: "tmp/test",
	}

	opt := packOptions(t, &gocrx.Options{
		Sign: true,
	})

	tc1, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.NotNil(t, tc1.Manifest):
		fallthrough
	case !assert.NotEmpty(t, tc1.PEM):
		fallthrough
	case !assert.NotEmpty(t, tc1.CRX):
		return
	}

	tcf := filepath.Join(dir, "crx")

	if !assert.NoError(t, writeBytesToFile(tcf, tc1.CRX)) {
		return
	}

	tpf := filepath.Join(dir, "pem")

	if !assert.NoError(t, writeBytesToFile(tpf, []byte(tc1.PEM))) {
		return
	}

	inp = &gocrx.Input{
		Packer:    gocrx.GoCRX3,
		Directory: "tmp/test",
		PEM:       tpf,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc2, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc2.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc2.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc2.CRX)):
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.GoCRX3,
		Binary: tc1.CRX,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc3, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc3.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc3.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc3.CRX)):
		return
	}

	tzf := filepath.Join(dir, "zip")

	if !assert.NoError(t, crx3.ZipTo(tzf, "tmp/test")) {
		return
	}

	bin, err := os.ReadFile(tzf)

	if !assert.NoError(t, err) {
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.GoCRX3,
		Binary: bin,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc4, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc4.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc4.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc4.CRX)):
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.GoCRX3,
		File:   tcf,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc5, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc5.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc5.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc5.CRX)):
		return
	}

	inp = &gocrx.Input{
		Packer: gocrx.GoCRX3,
		File:   tzf,
		PEM:    tc1.PEM,
	}

	opt = packOptions(t, &gocrx.Options{})

	tc6, err := gocrx.Pack(ctx, inp, opt)

	switch {
	case !assert.NoError(t, err):
		fallthrough
	case !assert.Equal(t, tc1.Manifest, tc6.Manifest):
		fallthrough
	case !assert.Equal(t, tc1.PEM, tc6.PEM):
		fallthrough
	case !assert.True(t, crxEqual(tc1.CRX, tc6.CRX)):
		return
	}
}

func packOptions(t *testing.T, opt *gocrx.Options) *gocrx.Options {
	opt.OnError = func(_ context.Context, err error) {
		assert.NoError(t, err)
	}

	return opt
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

func crxEqual(a []byte, b []byte) bool {
	return bytes.Equal(a[:bytes.IndexByte(a, 0x00)], b[:bytes.IndexByte(b, 0x00)])
}
