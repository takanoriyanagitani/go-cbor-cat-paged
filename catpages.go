package catpages

import (
	"context"
	"errors"
	"iter"
	"log"
)

var (
	ErrInvalidPage     error = errors.New("invalid page")
	ErrInvalidPageSize error = errors.New("invalid page size")
)

type PageSize uint32

const (
	PageSizeTiny   PageSize = 8
	PageSizeSmall  PageSize = 64
	PageSizeNormal PageSize = 512
	PageSizeLarge  PageSize = 4096
	PageSizeLARGE  PageSize = 32768
	PageSizeHuge   PageSize = 262144
	PageSizeHUGE   PageSize = 2097152
)

var PageSizeNameMap map[string]PageSize = map[string]PageSize{
	"Tiny":   PageSizeTiny,
	"Small":  PageSizeSmall,
	"Normal": PageSizeNormal,
	"Large":  PageSizeLarge,
	"LARGE":  PageSizeLARGE,
	"Huge":   PageSizeHuge,
	"HUGE":   PageSizeHUGE,
}

func PageSizeFromString(psname string) (PageSize, error) {
	sz, ok := PageSizeNameMap[psname]
	switch ok {
	case true:
		return sz, nil
	default:
		return PageSizeTiny, ErrInvalidPageSize
	}
}

type IgnoreError func(error) (ignore bool)
type LogError func(error)

func LogErrorSimple(e error) {
	log.Printf("%v\n", e)
}

var LogErrorDefault LogError = LogErrorSimple

type Page struct {
	size PageSize
	page []byte
	err  error
}

func (p Page) Err() error { return p.err }

func (p Page) Content() []byte { return p.page }

type PageFactory struct {
	PageSize
}

func PageFactoryNew(psz PageSize) PageFactory {
	return PageFactory{PageSize: psz}
}

func (f PageFactory) NewPage(content []byte, err error) Page {
	return Page{
		size: f.PageSize,
		page: content,
		err:  err,
	}
}

type CatPage func(context.Context, Page) error
type CatAll func(context.Context, iter.Seq[Page]) error

type PageSource func(context.Context) iter.Seq[Page]

func (c CatPage) ToCatAll(ie IgnoreError, le LogError) CatAll {
	return func(ctx context.Context, pages iter.Seq[Page]) error {
		for page := range pages {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			ep := page.Err()
			if nil != ep {
				le(ep)
				var ignore bool = ie(ep)
				if !ignore {
					return ep
				}

				continue
			}

			e := c(ctx, page)
			if nil != e {
				return e
			}
		}
		return nil
	}
}

func (a CatAll) ProcessAll(ctx context.Context, s PageSource) error {
	var i iter.Seq[Page] = s(ctx)
	return a(ctx, i)
}
