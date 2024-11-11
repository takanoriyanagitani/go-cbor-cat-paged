package rdr2page

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"

	cp "github.com/takanoriyanagitani/go-cbor-cat-paged"
	util "github.com/takanoriyanagitani/go-cbor-cat-paged/util"

	itools "github.com/takanoriyanagitani/go-cbor-cat-paged/util/iter"
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

func (s ReadSeeker) IdxToPageSource(idx iter.Seq[uint32]) cp.PageSource {
	return func(_ context.Context) iter.Seq[cp.Page] {
		var offsets iter.Seq[int64] = itools.Map(
			idx,
			func(ix uint32) int64 {
				var i int64 = int64(ix)
				i *= int64(s.PageFactory.PageSize)
				return i
			},
		)
		return func(yield func(cp.Page) bool) {
			var sz int = int(s.PageSize)
			var buf []byte = make([]byte, sz)
			for offset := range offsets {
				_, e := s.ReadSeeker.Seek(offset, io.SeekStart)
				if nil != e {
					se := fmt.Errorf("%w: %v", ErrSeek, e)
					yield(s.PageFactory.NewPage(nil, se))
					return
				}

				_, e = io.ReadFull(s.ReadSeeker, buf)
				if nil != e {
					if io.EOF == e {
						return
					}

					var page cp.Page = s.PageFactory.NewPage(
						buf,
						fmt.Errorf("%w: %v", ErrRead, e),
					)

					yield(page)
					return
				}

				var page cp.Page = s.PageFactory.NewPage(
					buf,
					nil,
				)

				if !yield(page) {
					return
				}
			}
		}
	}
}

func (s ReadSeeker) IdxToPageSourceInfinite() cp.PageSource {
	return s.IdxToPageSource(itools.Ints32u())
}

type IndexSource util.IO[iter.Seq[uint32]]

type ReaderSource util.IO[io.Reader]

func (r ReaderSource) ToIndexSource() util.IO[iter.Seq[uint32]] {
	return func(ctx context.Context) (iter.Seq[uint32], error) {
		rdr, e := r(ctx)
		if nil != e {
			return nil, e
		}
		return func(yield func(uint32) bool) {
			var buf [4]byte
			for {
				_, e := io.ReadFull(rdr, buf[:])
				if nil != e {
					return
				}
				var u5 uint32 = binary.BigEndian.Uint32(buf[:])
				if !yield(u5) {
					return
				}
			}
		}, nil
	}
}
