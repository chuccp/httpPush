package util

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type request struct {
	client *http.Client
}

func NewRequest() *request {
	ct := http.Client{Timeout: time.Second * 3}
	return &request{client: &ct}
}
func (r *request) Call(link string, jsonData []byte) ([]byte, error) {
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
func (r *request) JustCall(link string, jsonData []byte) error {
	var buff = new(bytes.Buffer)
	buff.Write(jsonData)
	_, err := r.client.Post(link, "application/json", buff)
	return err
}
