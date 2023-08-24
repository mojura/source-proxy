package proxy

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vroomy/httpserve"
)

func updateFilename(filename string) string {
	unix := time.Now().UnixNano()
	unixStr := strconv.FormatInt(unix, 10)
	return p.match.ReplaceAllString(filename, unixStr)
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
