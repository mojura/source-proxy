package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mojura/kiroku"
)

var _ kiroku.Source = &Client{}

func New(host, apiKey string) (cc *Client, err error) {
	var u *url.URL
	if u, err = url.Parse(host); err != nil {
		return
	}

	var c Client
	c.host = host
	c.apiKey = apiKey
	c.u = *u
	cc = &c
	return
}

type Client struct {
	hc http.Client

	u url.URL

	host   string
	apiKey string
}

func (c *Client) Get(ctx context.Context, prefix, filename string, fn func(io.Reader) error) (err error) {
	filename = url.PathEscape(filename)
	endpoint := fmt.Sprintf("/api/proxy/file/%s/%s", prefix, filename)
	err = c.request(ctx, "GET", endpoint, nil, fn)
	switch {
	case err == nil:
		return
	case strings.Contains(err.Error(), "NoSuchKey"):
		return os.ErrNotExist
	default:
		return
	}
}

// GetNext will get the next filename
// Note: prefix is ignored for source proxy as it derives the prefix from the filename
func (c *Client) GetNext(ctx context.Context, prefix, lastFilename string) (filename string, err error) {
	if lastFilename == "" {
		lastFilename = prefix
	}

	endpoint := fmt.Sprintf("/api/proxy/next/%s/%s", prefix, lastFilename)

	var resp apiResp
	resp.Data = &filename
	err = c.request(ctx, "GET", endpoint, nil, func(r io.Reader) (err error) {
		return json.NewDecoder(r).Decode(&resp)
	})

	switch {
	case err == nil:
		return
	case strings.Contains(err.Error(), "EOF"):
		err = io.EOF
		return
	default:
		return
	}
}

// GetNext will get the next filename
// Note: prefix is ignored for source proxy as it derives the prefix from the filename
func (c *Client) GetNextList(ctx context.Context, prefix, lastFilename string, maxKeys int64) (filenames []string, err error) {
	if lastFilename == "" {
		lastFilename = prefix
	}

	var resp apiResp
	resp.Data = &filenames
	endpoint := fmt.Sprintf("/api/proxy/nextList/%s/%s/%d", prefix, lastFilename, maxKeys)
	err = c.request(ctx, "GET", endpoint, nil, func(r io.Reader) (err error) {
		return json.NewDecoder(r).Decode(&resp)
	})

	switch {
	case err == nil:
		return
	case strings.Contains(err.Error(), "EOF"):
		err = io.EOF
		return
	default:
		return
	}
}

func (c *Client) GetInfo(ctx context.Context, prefix, filename string) (info kiroku.Info, err error) {
	filename = url.PathEscape(filename)
	endpoint := fmt.Sprintf("/api/proxy/info/%s/%s", prefix, filename)
	var resp apiResp
	resp.Data = &info
	err = c.request(ctx, "GET", endpoint, nil, func(r io.Reader) (err error) {
		return json.NewDecoder(r).Decode(&resp)
	})

	switch {
	case err == nil:
		return
	case strings.Contains(err.Error(), "NoSuchKey"):
		err = os.ErrNotExist
		return
	default:
		return
	}
}

func (c *Client) Import(ctx context.Context, prefix, filename string, w io.Writer) (err error) {
	return c.Get(ctx, prefix, filename, func(r io.Reader) (err error) {
		_, err = io.Copy(w, r)
		return
	})
}

func (c *Client) Export(ctx context.Context, prefix, filename string, r io.Reader) (newFilename string, err error) {
	endpoint := fmt.Sprintf("/api/proxy/%s/%s", prefix, filename)

	err = c.request(ctx, "POST", endpoint, r, func(r io.Reader) (err error) {
		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, r)
		newFilename = buf.String()
		return
	})

	return
}

func (c *Client) request(ctx context.Context, method, endpoint string, body io.Reader, fn func(r io.Reader) error) (err error) {
	u := c.u
	u.Path = endpoint
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, method, u.String(), body); err != nil {
		return
	}

	req.Header.Add("X-Api-Key", c.apiKey)

	var res *http.Response
	if res, err = c.hc.Do(req); err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return handleError(res)
	}

	if fn == nil {
		return
	}

	return fn(res.Body)
}
