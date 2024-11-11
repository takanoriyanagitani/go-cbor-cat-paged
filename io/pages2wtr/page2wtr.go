package page2wtr

import (
	"context"
	"io"

	cp "github.com/takanoriyanagitani/go-cbor-cat-paged"
)

type Writer struct {
	io.Writer
}

func (w Writer) ToCatPage() cp.CatPage {
	return func(_ context.Context, p cp.Page) error {
		_, e := w.Writer.Write(p.Content())
		return e
	}
}
