package rdr2pages2wtr

import (
	"context"

	cp "github.com/takanoriyanagitani/go-cbor-cat-paged"

	iw "github.com/takanoriyanagitani/go-cbor-cat-paged/io/pages2wtr"
	ir "github.com/takanoriyanagitani/go-cbor-cat-paged/io/rdr2pages"
)

type App struct {
	ir.ReadSeeker
	iw.Writer
	cp.IgnoreError
	cp.LogError
}

func (a App) ToCatPage() cp.CatPage       { return a.Writer.ToCatPage() }
func (a App) ToPageSource() cp.PageSource { return a.ReadSeeker.ToPageSource() }

func (a App) ToCatAll() cp.CatAll {
	return a.ToCatPage().ToCatAll(
		a.IgnoreError,
		a.LogError,
	)
}

func (a App) ProcessAll(ctx context.Context) error {
	return a.ToCatAll().ProcessAll(
		ctx,
		a.ToPageSource(),
	)
}
