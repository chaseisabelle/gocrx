package gocrx_test

import (
	"context"
	"github.com/chaseisabelle/gocrx"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPack(t *testing.T) {
	t.Parallel()

	inp := gocrx.Input{
		Packer:    gocrx.Chrome,
		Directory: "tmp/test",
	}

	opt := gocrx.Options{
		Sign: true,
	}

	out, err := gocrx.Pack(context.Background(), inp, opt)

	assert.NoError(t, err)
	assert.NotNil(t, out.Manifest)
	assert.NotNil(t, out.PEMBytes)
	assert.NotNil(t, out.CRXBytes)

	man := out.Manifest
	pem := out.PEMBytes
	crx := out.CRXBytes

	err = os.WriteFile("tmp/test.crx", out.CRXBytes, 0644)

	assert.NoError(t, err)

	err = os.WriteFile("tmp/test.pem", out.PEMBytes, 0644)

	assert.NoError(t, err)

	inp = gocrx.Input{
		Packer:    gocrx.Chrome,
		Directory: "tmp/test",
		PEMFile:   "tmp/test.pem",
	}

	opt = gocrx.Options{}

	out, err = gocrx.Pack(context.Background(), inp, opt)

	assert.NoError(t, err)
	assert.Equal(t, man, out.Manifest)
	assert.Equal(t, pem, out.PEMBytes)
	assert.Equal(t, crx, out.CRXBytes)

	ver := "420"

	inp = gocrx.Input{
		Packer:    gocrx.Chrome,
		Directory: "tmp/test",
		PEMFile:   "tmp/test.pem",
	}

	opt = gocrx.Options{
		Version: ver,
	}

	out, err = gocrx.Pack(context.Background(), inp, opt)

	assert.NoError(t, err)
	assert.NotNil(t, out.Manifest)
	assert.NotNil(t, out.PEMBytes)
	assert.NotNil(t, out.CRXBytes)

	assert.Equal(t, ver, out.Manifest["version"])
	assert.Equal(t, pem, out.PEMBytes)
	//assert.NotEqual(t, crx, out.CRXBytes) //<< @todo wtf?

	inp = gocrx.Input{
		Packer:    gocrx.GoCRX3,
		Directory: "tmp/test",
		PEMFile:   "tmp/test.pem",
	}

	opt = gocrx.Options{}

	out, err = gocrx.Pack(context.Background(), inp, opt)

	assert.NoError(t, err)
	assert.Equal(t, man, out.Manifest)
	assert.Equal(t, pem, out.PEMBytes)
	//assert.Equal(t, crx, out.CRXBytes) //<< @todo does not pack the same hmm...

	ver = "69"

	inp = gocrx.Input{
		Packer:    gocrx.GoCRX3,
		Directory: "tmp/test",
		PEMFile:   "tmp/test.pem",
	}

	opt = gocrx.Options{
		Version: ver,
	}

	out, err = gocrx.Pack(context.Background(), inp, opt)

	assert.NoError(t, err)
	assert.NotNil(t, out.Manifest)
	assert.NotNil(t, out.PEMBytes)
	assert.NotNil(t, out.CRXBytes)

	assert.Equal(t, ver, out.Manifest["version"])
	assert.Equal(t, pem, out.PEMBytes)
	assert.NotEqual(t, crx, out.CRXBytes)
}
