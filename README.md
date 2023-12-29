# gocrx
***pack chrome extensions with go***

---
## install

```bash
go get github.com/chaseisabelle/gocrx
```

---
## usage

```go
import "github.com/chaseisabelle/gocrx"
```

```go
// pack a new crx and generate a pem

inp := gocrx.Input{
    Packer:    gocrx.Chrome, 
    Directory: "/path/to/extension",
}

opt := gocrx.Options{
    Sign: true,
}

out, err := gocrx.Pack(context.Background(), inp, opt)

os.WriteFile("/path/to/extension.crx", out.CRXBytes, 0644)
os.WriteFile("/path/to/extension.pem", out.PEMBytes, 0644)
```

```go
// pack a crx with an existing pem

inp := gocrx.Input{
    Packer:    gocrx.Chrome, 
    Directory: "/path/to/extension",
    PEMFile:   "/path/to/extension.pem",
}

opt := gocrx.Options{}

out, err := gocrx.Pack(context.Background(), inp, opt)

os.WriteFile("/path/to/extension.crx", out.CRXBytes, 0644)
```

---
## packers

- `Chrome` is the default packer
  - pro: it uses the chrome/chromium binary
  - con: it **requires** chrome/chromium to be installed
- `GoCRX3` the [go-crx3](https://github.com/mmadfox/go-crx3) packer
  - pro: it's a pure go implementation
  - con: it _might not_ be as reliable

---
## some things to think about...

- this package is pretty heavy on disk io
- it is not very efficient/performant
- it's intent was to be used for ci/cd
- the go-crx3 packer is kinda/sorta experimental
- the go-crx3 packer cannot sign crx files (yet)