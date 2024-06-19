package test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

const (
	testPort        = 8080
	testDocRoot     = "testdata/htdocs"
	serverSetupTime = 1 // second

	contentTypeHTML = "text/html; charset=utf-8"
	contentTypeJPG  = "image/jpeg"
	contentTypePNG  = "image/png"
)

type concurrentTestSpec struct {
	// reqPath is the path to the request file to send.
	reqPath string

	resChecker *ResponseChecker
}

// To test concurrent request handling, we send 2 requests.
// The first one doesn't have the "Connection: close" header,
// so it would hang until server timeout (~5s).
// At the same time, we send a second request to the server.
// We check the responses from both reqeusts as a verification.
func TestConcurrentRequest(t *testing.T) {
	var tests = []struct {
		name  string
		specs []*concurrentTestSpec
	}{
		{
			"OKOK",
			[]*concurrentTestSpec{
				{
					"testdata/requests/single/OKTimeout.txt",
					&ResponseChecker{
						StatusCode:  200,
						FilePath:    filepath.Join(testDocRoot, "index.html"),
						ContentType: contentTypeHTML,
						Close:       false,
					},
				},
				{
					"testdata/requests/single/OKBasic.txt",
					&ResponseChecker{
						StatusCode:  200,
						FilePath:    filepath.Join(testDocRoot, "index.html"),
						ContentType: contentTypeHTML,
						Close:       true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start test server
			serverCmd := exec.Command(
				"_bin/httpd",
				"-port", strconv.Itoa(testPort),
				"-doc_root", testDocRoot,
			)
			if err := serverCmd.Start(); err != nil {
				log.Fatal(err)
			}
			// Wait a little bit for the server to set up.
			// Otherwise we might get "connection refused".
			time.Sleep(serverSetupTime * time.Second)
			// Kill the test server process
			defer serverCmd.Process.Kill()

			errChan := make(chan error)
			for i, spec := range tt.specs {
				reqPath := spec.reqPath
				resPath := filepath.Join("testdata/responses/concurrent", fmt.Sprintf("%v%v.dat", tt.name, i))

				// Send requests and check responses concurrently
				go func(spec *concurrentTestSpec) {
					defer func() {
						if r := recover(); r != nil {
							errChan <- fmt.Errorf("recovered from panic: %v", r)
						}
					}()

					c := &Client{Port: testPort}
					defer c.Close()
					if err := c.Dial(); err != nil {
						errChan <- err
						return
					}
					if err := c.SendRequestFromFile(reqPath); err != nil {
						errChan <- err
						return
					}
					if err := c.ReceiveResponseToFile(resPath); err != nil {
						errChan <- err
						return
					}

					f, err := os.Open(resPath)
					if err != nil {
						errChan <- err
						return
					}
					br := bufio.NewReader(f)
					if err := spec.resChecker.Check(br); err != nil {
						errChan <- err
						return
					}
					if _, err := br.ReadByte(); !errors.Is(err, io.EOF) {
						errChan <- fmt.Errorf("response has extra bytes when it should end")
						return
					}
					errChan <- nil
				}(spec)
			}

			for range tt.specs {
				err := <-errChan
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
