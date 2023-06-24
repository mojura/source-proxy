package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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

func (c *Client) Get(ctx context.Context, filename string, fn func(io.Reader) error) (err error) {
	endpoint := fmt.Sprintf("/api/proxy/file/%s", filename)
	err = c.request(ctx, "GET", endpoint, nil, fn)
	return
}

func (c *Client) GetNext(ctx context.Context, lastFilename string) (filename string, err error) {
	endpoint := fmt.Sprintf("/api/proxy/next/%s", lastFilename)
	var resp apiResp
	resp.Data = &filename
	if err = c.request(ctx, "GET", endpoint, nil, func(r io.Reader) (err error) {
		return json.NewDecoder(r).Decode(&resp)
	}); err != nil {
		return
	}

	if len(resp.Errors) > 0 {
		err = errors.New(strings.Join(resp.Errors, "\n"))
		return
	}

	return
}

func (c *Client) Import(ctx context.Context, filename string, w io.Writer) (err error) {
	return c.Get(ctx, filename, func(r io.Reader) (err error) {
		_, err = io.Copy(w, r)
		return
	})
}

func (c *Client) Export(ctx context.Context, filename string, r io.Reader) (err error) {
	endpoint := fmt.Sprintf("/api/proxy/%s", filename)
	err = c.request(ctx, "POST", endpoint, r, nil)
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

func handleError(res *http.Response) (err error) {
	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, res.Body); err != nil {
		return
	}

	return fmt.Errorf("error encountered (%d): %v", res.StatusCode, buf.String())
}

type apiResp struct {
	Data   interface{} `json:"data"`
	Errors []string    `json:"errors"`
}
