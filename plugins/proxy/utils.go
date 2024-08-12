package proxy

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mojura/kiroku"
	"github.com/vroomy/httpserve"
)

func updateFilename(filename string) string {
	var (
		f   kiroku.Filename
		err error
	)

	if f, err = kiroku.ParseFilename(filename); err != nil {
		return filename
	}

	unix := time.Now().UnixNano()
	f.CreatedAt = unix
	return f.String()
}

func getAPIKey(ctx *httpserve.Context) (apikey string, err error) {
	var (
		vals []string
		ok   bool
	)

	if vals, ok = ctx.Request().Header["X-Api-Key"]; !ok || len(vals) == 0 {
		err = errors.New("header field of <X-Api-Key> is uset")
		return
	}

	apikey = vals[0]
	return
}

func getResource(prefix, filename string) (resource string, err error) {
	if prefix != "_latestSnapshots" {
		resource = prefix
		return
	}

	resource = strings.Replace(filename, "_latestSnapshots/", "", 1)
	partEnd := strings.Index(filename, ".")
	if partEnd == -1 {
		err = fmt.Errorf("invalid filename <%s> does not have a part separator", filename)
		return
	}

	if resource, err = url.PathUnescape(filename[0:partEnd]); err != nil {
		err = fmt.Errorf("error unescaping filename: %v", err)
		return
	}

	return
}

func copyToTemp(r io.Reader, fn func(f *os.File) error) (err error) {
	var f *os.File
	if f, err = os.CreateTemp("", "exporting"); err != nil {
		err = fmt.Errorf("error creating temporary file: %v", err)
		return
	}
	name := f.Name()
	defer os.Remove(name)
	defer f.Close()

	if _, err = io.Copy(f, r); err != nil {
		err = fmt.Errorf("error copying to temporary file: %v", err)
		return
	}

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		err = fmt.Errorf("error seeking within temporary file: %v", err)
		return
	}

	return fn(f)
}
