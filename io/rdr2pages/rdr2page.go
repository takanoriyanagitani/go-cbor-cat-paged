package rdr2page

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"

	cp "github.com/takanoriyanagitani/go-cbor-cat-paged"
)

var (
	ErrRead error = errors.New("error while reading")
	ErrSeek error = errors.New("error while seeking")
)

type ReadSeeker struct {
	io.ReadSeeker
	cp.PageFactory
}

func (s ReadSeeker) ToPageSource() cp.PageSource {
	return func(_ context.Context) iter.Seq[cp.Page] {
		return func(yield func(cp.Page) bool) {
			var sz uint32 = uint32(s.PageFactory.PageSize)
			var buf []byte = make([]byte, sz)
			var offset int64 = 0
			for {
				_, re := io.ReadFull(s.ReadSeeker, buf)
				if nil != re {
					if io.EOF == re {
						return
					}
					re = fmt.Errorf("%w: %v", ErrRead, re)
				}

				offset += int64(sz)

				_, se := s.ReadSeeker.Seek(offset, io.SeekStart)
				if nil != se {
					se = fmt.Errorf("%w: %v", ErrSeek, se)
				}

				page := s.PageFactory.NewPage(
					buf,
					errors.Join(re, se),
				)

				if !yield(page) {
					return
				}
			}
		}
	}
}

func IgnoreReadError(e error) (ignore bool) {
	return errors.Is(e, ErrRead)
}

var IgnoreErrDefault cp.IgnoreError = IgnoreReadError
