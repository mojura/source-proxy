package proxy

import "testing"

func Test_getResource(t *testing.T) {
	type args struct {
		prefix   string
		filename string
	}

	type testcase struct {
		name         string
		args         args
		wantResource string
		wantErr      bool
	}

	tests := []testcase{
		{
			name: "basic",
			args: args{
				prefix:   "myprefix",
				filename: "testing.12345.snapshot.kir",
			},
			wantResource: "testing",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResource, err := getResource(tt.args.prefix, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("getResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotResource != tt.wantResource {
				t.Errorf("getResource() = %v, want %v", gotResource, tt.wantResource)
			}
		})
	}
}
