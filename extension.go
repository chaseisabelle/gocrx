package gocrx

import "context"

type Extension interface {
	Manifest(context.Context) (Manifest, error)
	Close(context.Context) error
}
