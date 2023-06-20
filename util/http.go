package util

import (
	"net/http"
	"strconv"
)

func NewServer() *HttpServer {
	return &HttpServer{serveMux: http.NewServeMux()}
}

type HttpServer struct {
	serveMux *http.ServeMux
}

func (hs *HttpServer) AddRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	hs.serveMux.HandleFunc(pattern, handler)
}
func (hs *HttpServer) Start(port int) error {
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: hs.serveMux,
	}
	error := srv.ListenAndServe()
	return error
}
func (hs *HttpServer) StartTLS(port int, certFile, keyFile string) error {
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: hs.serveMux,
	}
	return srv.ListenAndServeTLS(certFile, keyFile)
}
func (hs *HttpServer) StartAutoTLS(port int, certFile, keyFile string) error {
	if len(certFile) > 0 && len(keyFile) > 0 {
		return hs.StartTLS(port, certFile, keyFile)
	} else {
		return hs.Start(port)
	}
}
