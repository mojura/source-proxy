package client

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

var testclient *Client

func TestMain(m *testing.M) {
	var err error
	host := os.Getenv("SOURCE_PROXY_HOST")
	apiKey := os.Getenv("SOURCE_PROXY_APIKEY")
	if testclient, err = New(host, apiKey); err != nil {
		log.Fatal(err)
	}

	sc := m.Run()
	os.Exit(sc)
}

func TestClient_Export(t *testing.T) {
	type args struct {
		ctx      context.Context
		prefix   string
		filename string
		r        io.Reader
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				ctx:      context.Background(),
				prefix:   "test_prefix",
				filename: "bids.0000.chunk.moj",
				r:        strings.NewReader("foo bar baz"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := testclient.Export(tt.args.ctx, tt.args.prefix, tt.args.filename, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("Client.Export() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_GetNext(t *testing.T) {
	type args struct {
		ctx      context.Context
		filename string
		r        io.Reader
	}

	tests := []struct {
		name           string
		args           args
		noWantFilename string
		wantErr        bool
	}{
		{
			name: "basic",
			args: args{
				ctx:      context.Background(),
				filename: "bids.0000.chunk.moj",
				r:        strings.NewReader("foo bar baz"),
			},
			noWantFilename: "bids.0000.chunk.moj",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got string
				err error
			)

			if got, err = testclient.GetNext(tt.args.ctx, "", tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("Client.GetNext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == "" {
				t.Error("Client.GetNext() received empty string when it wasn't expected")
				return
			}

			if got == tt.noWantFilename {
				t.Errorf("Client.GetNext() received %s when it wasn't expected", tt.noWantFilename)
				return
			}
		})
	}
}
