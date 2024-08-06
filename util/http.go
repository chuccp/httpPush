package util

import (
	"net/http"
	"strconv"
	"time"
)

func NewServer() *HttpServer {
	return &HttpServer{serveMux: http.NewServeMux()}
}

type HttpServer struct {
	serveMux *http.ServeMux
	isTls    bool
}

func (hs *HttpServer) IsTls() bool {
	return hs.isTls
}
func (hs *HttpServer) AddRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	hs.serveMux.HandleFunc(pattern, handler)
}

const MaxHeaderBytes = 8192

const MaxReadHeaderTimeout = time.Second * 30

const MaxReadTimeout = time.Minute * 2

const MaxIdleTimeout = time.Minute * 5

func (hs *HttpServer) Start(port int) error {
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           hs.serveMux,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
		IdleTimeout:       MaxIdleTimeout,
	}
	hs.isTls = false
	error := srv.ListenAndServe()
	return error
}
func (hs *HttpServer) StartTLS(port int, certFile, keyFile string) error {
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           hs.serveMux,
		ReadHeaderTimeout: MaxReadHeaderTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ReadTimeout:       MaxReadTimeout,
		IdleTimeout:       MaxIdleTimeout,
	}
	hs.isTls = true
	return srv.ListenAndServeTLS(certFile, keyFile)
}
func (hs *HttpServer) StartAutoTLS(port int, certFile, keyFile string) error {
	if len(certFile) > 0 && len(keyFile) > 0 {
		return hs.StartTLS(port, certFile, keyFile)
	} else {
		return hs.Start(port)
	}
}
