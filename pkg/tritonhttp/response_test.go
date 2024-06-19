package tritonhttp

import (
	"bytes"
	"testing"
)

func TestWriteSortedHeaders(t *testing.T) {
	var tests = []struct {
		name string
		res  *Response
		want string
	}{
		{
			"Basic",
			&Response{
				Header: map[string]string{
					"Connection": "close",
					"Date":       "foobar",
					"Misc":       "hello world",
				},
			},
			"Connection: close\r\n" +
				"Date: foobar\r\n" +
				"Misc: hello world\r\n" +
				"\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("recovered from panic: %v", r)
				}
			}()
			var buffer bytes.Buffer
			if err := tt.res.WriteSortedHeaders(&buffer); err != nil {
				t.Fatal(err)
			}
			got := buffer.String()
			if got != tt.want {
				t.Fatalf("got: %q, want: %q", got, tt.want)
			}
		})
	}
}
