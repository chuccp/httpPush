package util

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
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

type netBreak struct {
	isBreak   bool
	breakTime time.Time
	lock      sync.Mutex
	err       error
	num       int32
}

func (nb *netBreak) beBreak(err error) {
	nb.lock.Lock()
	defer nb.lock.Unlock()
	nb.isBreak = true
	nb.breakTime = time.Now()
	nb.err = err
}
func (nb *netBreak) noBreak() {
	if nb.isBreak {
		nb.lock.Lock()
		defer nb.lock.Unlock()
		nb.isBreak = false
	}
}
func (nb *netBreak) hasBreak() (error, bool) {
	nb.lock.Lock()
	defer nb.lock.Unlock()
	ti := time.Now().Add(time.Second * -5)
	if ti.After(nb.breakTime) {
		return nil, false
	}
	return nb.err, nb.isBreak
}

type Request struct {
	client   *http.Client
	netBreak *netBreak
}

func NewRequest() *Request {
	//retryTransport := &RetryTransport{
	//	Transport: http.DefaultTransport,
	//	Retries:   0,
	//}
	ct := http.Client{Timeout: time.Second * 1, Transport: http.DefaultTransport}
	return &Request{client: &ct, netBreak: &netBreak{isBreak: false}}
}

func (r *Request) CallBreak(link string, jsonData []byte) ([]byte, error) {
	err, b := r.netBreak.hasBreak()
	if b {
		return nil, err
	}
	if r.netBreak.isBreak {
		if !atomic.CompareAndSwapInt32(&r.netBreak.num, 0, 1) {
			return nil, r.netBreak.err
		}
	}
	call, err := r.Call(link, jsonData)
	atomic.StoreInt32(&r.netBreak.num, 0)
	if err != nil {
		if strings.Contains(err.Error(), "No connection could be made because the target machine actively refused it") {
			r.netBreak.beBreak(err)
		}
		return nil, err
	}
	return call, nil
}
func (r *Request) Call(link string, jsonData []byte) ([]byte, error) {
	var buff = new(bytes.Buffer)
	buff.Write(jsonData)
	resp, err := r.client.Post(link, "application/json", buff)
	if err != nil {
		return nil, err
	}
	r.netBreak.noBreak()
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
	r.netBreak.noBreak()
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
	if err == nil {
		r.netBreak.noBreak()
	}
	return err
}
