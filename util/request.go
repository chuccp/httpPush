package util

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type RetryTransport struct {
	Transport http.RoundTripper
	Retries   int
}

func (r *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i <= r.Retries; i++ {
		resp, err = r.Transport.RoundTrip(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return resp, err
}

type Request struct {
	client *http.Client
}

func NewRequest() *Request {
	retryTransport := &RetryTransport{
		Transport: http.DefaultTransport,
		Retries:   3,
	}
	ct := http.Client{Timeout: time.Second * 10, Transport: retryTransport}
	return &Request{client: &ct}
}
func (r *Request) Call(link string, jsonData []byte) ([]byte, error) {
	var buff = new(bytes.Buffer)
	buff.Write(jsonData)
	resp, err := r.client.Post(link, "application/json", buff)
	if err != nil {
		return nil, err
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return all, nil
}
func (r *Request) Get(link string) ([]byte, error) {
	resp, err := r.client.Get(link)
	if err != nil {
		return nil, err
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return all, nil
}
func (r *Request) JustCall(link string, jsonData []byte) error {
	var buff = new(bytes.Buffer)
	buff.Write(jsonData)
	_, err := r.client.Post(link, "application/json", buff)
	return err
}
