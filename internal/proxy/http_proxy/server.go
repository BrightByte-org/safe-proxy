package http_proxy

import (
	"encoding/json"
	"errors"
	"github.com/iooojik/go-logger"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

const (
	HeaderContentType          = "Content-Type"
	HeaderContentTypeValueJson = "application/json"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

type Status struct {
	Message string
	Code    int
}

type HttpProxy struct {
	Addr string
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func sendError(err error, status int, w http.ResponseWriter) {
	w.Header().Add(HeaderContentType, HeaderContentTypeValueJson)
	w.WriteHeader(status)
	logger.LogError(err)
	e := json.NewEncoder(w).Encode(Status{
		Message: err.Error(),
		Code:    status,
	})
	if e != nil {
		//todo replace
		panic(e)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

func (p *HttpProxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		msg := "unsupported protocal scheme " + req.URL.Scheme
		sendError(errors.New(msg), http.StatusBadRequest, wr)
		return
	}
	req.RequestURI = ""
	delHopHeaders(req.Header)
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		sendError(err, http.StatusBadRequest, wr)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			sendError(err, http.StatusBadRequest, wr)
		}
	}(resp.Body)
	log.Println(req.RemoteAddr, " ", resp.Status)
	delHopHeaders(resp.Header)
	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	_, err = io.Copy(wr, resp.Body)
	if err != nil {
		sendError(err, http.StatusBadRequest, wr)
	}
}
