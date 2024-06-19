package tritonhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	resProto = "HTTP/1.1"

	statusOK = 200
	statusBadReq = 400
	statusNotFound = 404
)

var statusText = map[int]string{
	statusOK: "OK",
	statusBadReq: "Bad Request",
	statusNotFound: "Not Found",
}

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	DocRoot string
}

func (s *Server) ValidateServer() error {
	fi, err := os.Stat(s.DocRoot)

	if os.IsNotExist(err) {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("doc root %q is not a directory", s.DocRoot)
	}

	return nil
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	if err := s.ValidateServer(); err != nil {
		return err
	}

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer l.Close()

	// Hint: call HandleConnection
	for{
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go s.HandleConnection(conn)
	}
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	// Hint: use the other methods below

	for {
		// Set timeout
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			log.Printf("Failed to set Timeout for connection %v", conn)
			_ = conn.Close()
			return
		}

		// Try to read next request
		req, bytesReceived, err := ReadRequest(reader)

		// handle errors
		// error 1: client close the conn -> io.EOF error
		if errors.Is(err, io.EOF) {
			_ = conn.Close()
			return
		}

		// error 2: timeout from the server -> net.Error
		if err, ok := err.(net.Error); ok && err.Timeout() {
			if !bytesReceived {
				_ = conn.Close()
				return
			}else {
				res := &Response{}
				res.HandleBadRequest()
				_ = res.Write(conn)
				_ = conn.Close()
				return
			}
		}

		// error 3: invalid request
		if err != nil {
			res := &Response{}
			res.HandleBadRequest()
			_ = res.Write(conn)
			_ = conn.Close()
			return
		}

		// Handle good request
		res := s.HandleGoodRequest(req)
		err = res.Write(conn)
		if err != nil {
			fmt.Println(err)
		}

		// Close conn if requested
		if req.Close {
			conn.Close()
			break
		}
	}
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	// Hint: use the other methods below
	res = &Response{}

	if strings.HasSuffix(req.URL, "/") {
		req.URL = req.URL + "index.html"
	}

	path := filepath.Clean(filepath.Join(s.DocRoot, req.URL))
	
	rel, _ := filepath.Rel(s.DocRoot, path)
	if strings.HasPrefix(rel, "../") {
		res.HandleNotFound(req)
		return res
	}

	f, err := os.Stat(path)
	if err != nil || f.IsDir(){
		res.HandleNotFound(req)
		return res
	}

	res.HandleOK(req, path)

	return res
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {
	fiInfo, err := os.Stat(path)
	if err != nil {
		log.Panicln(err)
	}
	okHeaderMap := make(map[string]string)
	res.Proto = resProto
	res.StatusCode = statusOK
	res.FilePath = path
	res.Request = req

	okHeaderMap["Date"] = FormatTime(time.Now())
	okHeaderMap["Last-Modified"] = FormatTime(fiInfo.ModTime())
	okHeaderMap["Content-Type"] = MIMETypeByExtension(filepath.Ext(path))
	okHeaderMap["Content-Length"] = strconv.FormatInt(fiInfo.Size(),10)
	if req.Close {
		okHeaderMap["Connection"] = "close"
	}

	res.Header = okHeaderMap
}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	badReqHeaderMap := make(map[string]string)
	res.Proto = resProto
	res.StatusCode = statusBadReq
	res.FilePath = ""
	res.Request = nil

	badReqHeaderMap["Date"] = FormatTime(time.Now())
	badReqHeaderMap["Connection"] = "close"

	res.Header = badReqHeaderMap
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {
	NotFoundHeaderMap := make(map[string]string)
	res.Proto = resProto
	res.StatusCode = statusNotFound
	res.FilePath = ""
	res.Request = req

	NotFoundHeaderMap["Date"] = FormatTime(time.Now())
	if req.Close {
		NotFoundHeaderMap["Connection"] = "close"
	}

	res.Header = NotFoundHeaderMap
}
