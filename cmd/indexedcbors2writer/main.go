package main

import (
	"bufio"
	"context"
	"io"
	"iter"
	"log"
	"os"

	cp "github.com/takanoriyanagitani/go-cbor-cat-paged"
	util "github.com/takanoriyanagitani/go-cbor-cat-paged/util"

	iw "github.com/takanoriyanagitani/go-cbor-cat-paged/io/pages2wtr"
	ir "github.com/takanoriyanagitani/go-cbor-cat-paged/io/rdr2pages"

	ap "github.com/takanoriyanagitani/go-cbor-cat-paged/app/idx2pages2wtr"
)

var logErr cp.LogError = cp.LogErrorSimple
var ignErr cp.IgnoreError = ir.IgnoreReadError

func GetEnvByKeyNew(key string) util.IO[string] {
	return func(_ context.Context) (val string, e error) {
		val = os.Getenv(key)
		return val, nil
	}
}

func GetPageSizeFromEnvNew(key string) util.IO[cp.PageSize] {
	return util.ComposeIoErr(
		GetEnvByKeyNew(key),
		cp.PageSizeFromString,
	)
}

func PageFactoryFromEnvNew(key string) util.IO[cp.PageFactory] {
	return util.ComposeIo(
		GetPageSizeFromEnvNew(key),
		cp.PageFactoryNew,
	)
}

type IoConfig struct {
	io.ReadSeekCloser
	io.Writer
}

func (i IoConfig) Close() error {
	return i.ReadSeekCloser.Close()
}

func (i IoConfig) ToReadSeeker(pf cp.PageFactory) ir.ReadSeeker {
	return ir.ReadSeeker{
		ReadSeeker:  i.ReadSeekCloser,
		PageFactory: pf,
	}
}

func (i IoConfig) ToWriter() iw.Writer { return iw.Writer{Writer: i.Writer} }

func (i IoConfig) ToApp(pf cp.PageFactory, idx iter.Seq[uint32]) ap.App {
	return ap.App{
		ReadSeeker:  i.ToReadSeeker(pf),
		Writer:      i.ToWriter(),
		IgnoreError: ignErr,
		LogError:    logErr,
		Index:       idx,
	}
}

func file2wtr(
	ctx context.Context,
	file io.ReadSeekCloser,
	w io.Writer,
	pf cp.PageFactory,
	isrc util.IO[iter.Seq[uint32]],
) error {
	icfg := IoConfig{
		ReadSeekCloser: file,
		Writer:         w,
	}
	defer icfg.Close()

	idx, e := isrc(ctx)
	if nil != e {
		return e
	}

	var a ap.App = icfg.ToApp(pf, idx)
	return a.ProcessAll(ctx)
}

func filename2stdout(
	ctx context.Context,
	filename string,
	pf cp.PageFactory,
) error {
	var br io.Reader = bufio.NewReader(os.Stdin)
	var rsrc ir.ReaderSource = func(_ context.Context) (io.Reader, error) {
		return br, nil
	}
	var isrc util.IO[iter.Seq[uint32]] = rsrc.ToIndexSource()

	var bw *bufio.Writer = bufio.NewWriter(os.Stdout)
	defer bw.Flush()

	file, e := os.Open(filename)
	if nil != e {
		return e
	}
	return file2wtr(ctx, file, bw, pf, isrc)
}

func sub(ctx context.Context) error {
	var filename string = os.Getenv("ENV_PAGED_CBOR_FILENAME")
	var pf util.IO[cp.PageFactory] = PageFactoryFromEnvNew("ENV_CBOR_PAGE_SIZE")
	p, e := pf(ctx)
	if nil != e {
		return e
	}
	return filename2stdout(ctx, filename, p)
}

func main() {
	e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}