package cluster

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

type request struct {
	client *http.Client
	lock   *sync.Mutex
}

func NewRequest() *request {
	transport := &http.Transport{
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	ct := http.Client{Timeout: time.Second * 2, Transport: transport}
	return &request{client: &ct, lock: new(sync.Mutex)}
}

func (r *request) call(link string, jsonData []byte, ctx context.Context) ([]byte, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	var buff = new(bytes.Buffer)
	buff.Write(jsonData)
	req, err := http.NewRequestWithContext(ctx, "POST", link, buff)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return all, nil
}

type HttpClient struct {
	requests map[string]*request
	lock     *sync.RWMutex
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		requests: make(map[string]*request),
		lock:     new(sync.RWMutex),
	}
}
func (client *HttpClient) getRequest(remoteAddress string) *request {
	client.lock.RLock()
	req, ok := client.requests[remoteAddress]
	if ok {
		client.lock.RUnlock()
		return req
	} else {
		client.lock.RUnlock()
		client.lock.Lock()
		req, ok := client.requests[remoteAddress]
		if ok {
			client.lock.Unlock()
			return req
		}
		req = NewRequest()
		client.requests[remoteAddress] = req
		client.lock.Unlock()
		return req
	}
}
func (client *HttpClient) Call(machine *Machine, path string, jsonData []byte) (data []byte, err error) {
	return client.CallByLink(machine.Link, path, jsonData)
}
func (client *HttpClient) CallByLink(link string, path string, jsonData []byte) (data []byte, err error) {
	req := client.getRequest(link)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	data, err = req.call(link+path, jsonData, ctx)
	return
}
