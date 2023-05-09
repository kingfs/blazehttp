package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/kingfs/blazehttp/http"
	progressbar "github.com/schollz/progressbar/v3"
)

var (
	target  string // the target web site, example: http://192.168.0.1:8080
	glob    string // use glob expression to select multi files
	timeout = 1000 // default 1000 ms
)

func init() {
	flag.StringVar(&target, "t", "", "target website, example: http://192.168.0.1:8080")
	flag.StringVar(&glob, "g", "", "glob expression, example: *.http")
	flag.IntVar(&timeout, "timeout", 1000, "connection timeout, default 1000 ms")
	flag.Parse()
}

func connect(addr string, isHttps bool, timeout int) net.Conn {
	var n net.Conn
	var err error

	retryCnt := 0
retry:
	if isHttps {
		n, err = tls.Dial("tcp", addr, nil)
	} else {
		n, err = net.Dial("tcp", addr)
	}
	if err != nil {
		retryCnt++
		if retryCnt < 4 {
			goto retry
		} else {
			return nil
		}
	}
	wDeadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
	rDeadline := time.Now().Add(time.Duration(timeout*2) * time.Millisecond)
	deadline := time.Now().Add(time.Duration(timeout*2) * time.Millisecond)
	n.SetDeadline(deadline)
	n.SetReadDeadline(rDeadline)
	n.SetWriteDeadline(wDeadline)

	return n
}

func main() {
	isHttps := false
	addr := target

	if strings.HasPrefix(target, "http") {
		u, _ := url.Parse(target)
		if u.Scheme == "https" {
			isHttps = true
		}
		addr = u.Host
	}

	fileList, err := filepath.Glob(glob)
	if err != nil || len(fileList) == 0 {
		fmt.Printf("cannot find http file")
		return
	}

	stats := make(map[int][]string)
	success := 0

	bar := progressbar.Default(int64(len(fileList)), "sending")
	for _, f := range fileList {
		bar.Add(1)
		req := new(http.Request)
		if err = req.ReadFile(f); err != nil {
			fmt.Printf("read request file: %s error: %s\n", f, err)
			continue
		}
		req.SetHost(addr)
		// one http request one connection
		req.SetHeader("Connection", "close")

		conn := connect(addr, isHttps, timeout)
		nWrite, err := req.WriteTo(conn)
		if err != nil {
			fmt.Printf("send request poc: %s length: %d error: %s", f, nWrite, err)
			continue
		}

		rsp := new(http.Response)
		if err = rsp.ReadConn(conn); err != nil {
			fmt.Printf("read poc file: %s response, error: %s", f, err)
			continue
		}
		success++
		statusCode := rsp.GetStatusCode()
		if _, ok := stats[statusCode]; !ok {
			stats[statusCode] = []string{f}
		} else {
			stats[statusCode] = append(stats[statusCode], f)
		}
	}

	fmt.Printf("Total http file: %d, success: %d failed: %d\n", len(fileList), success, (len(fileList) - success))

	for k, v := range stats {
		fmt.Printf("Status code: %d hit: %d\n", k, len(v))
	}

}
