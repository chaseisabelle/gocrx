package gocrx

import "context"

type Packer interface {
	PackAndSign(context.Context, Extension) ([]byte, string, error)
	PackWithPEM(context.Context, Extension, string) ([]byte, error)
}
