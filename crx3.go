package gocrx

import "context"

type CRX3 struct {
	logger Logger
}

func NewCRX3(lgr Logger) *CRX3 {
	if lgr == nil {
		lgr = func(context.Context, error) {}
	}

	return &CRX3{
		logger: lgr,
	}
}

func (c *CRX3) PackAndSign(ctx context.Context, ext Extension) ([]byte, string, error) {
	return nil, "", nil
}

func (c *CRX3) PackWithPEM(ctx context.Context, ext Extension, pem string) ([]byte, error) {
	return nil, nil
}
