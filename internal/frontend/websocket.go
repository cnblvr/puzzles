package frontend

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type websocketMessagesPool struct {
	mx           sync.Mutex
	requestPool  map[string]websocketRequest
	responsePool map[string]websocketResponse
}

func (p *websocketMessagesPool) Add(req websocketRequest, resp websocketResponse) {
	p.mx.Lock()
	defer p.mx.Unlock()
	req = reflect.New(reflect.TypeOf(req).Elem()).Interface().(websocketRequest)
	resp = reflect.New(reflect.TypeOf(resp).Elem()).Interface().(websocketResponse)
	p.requestPool[req.Method()] = req
	p.responsePool[resp.Method()] = resp
}

func (p *websocketMessagesPool) GetRequest(method string) (websocketRequest, error) {
	if method == "" {
		return nil, fmt.Errorf("method is empty")
	}
	p.mx.Lock()
	defer p.mx.Unlock()
	req, ok := p.requestPool[method]
	if !ok {
		return nil, fmt.Errorf("method not allowed")
	}
	return req, nil
}

var websocketPool = websocketMessagesPool{
	requestPool:  make(map[string]websocketRequest),
	responsePool: make(map[string]websocketResponse),
}

type websocketRequest interface {
	Method() string
	Validate(context.Context) error
	Execute(context.Context) (websocketResponse, error)
}

type websocketResponse interface {
	Method() string
	Validate(context.Context) error
	Execute(context.Context) error
}
