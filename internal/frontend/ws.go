package frontend

import (
	"context"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"reflect"
	"sync"
)

type wsMessagesPool struct {
	mx              sync.Mutex
	incomingReqPool map[string]wsIncomingRequest
}

func wsAddIncoming(method string, req wsIncomingRequest) {
	wsPool.mx.Lock()
	defer wsPool.mx.Unlock()
	req = reflect.New(reflect.TypeOf(req).Elem()).Interface().(wsIncomingRequest)
	wsPool.incomingReqPool[method] = req
}

func wsGetIncoming(method string) (wsIncomingRequest, error) {
	if method == "" {
		return nil, fmt.Errorf("method is empty")
	}
	wsPool.mx.Lock()
	defer wsPool.mx.Unlock()
	req, ok := wsPool.incomingReqPool[method]
	if !ok {
		return nil, fmt.Errorf("method not allowed")
	}
	return req, nil
}

var wsPool = wsMessagesPool{
	incomingReqPool: make(map[string]wsIncomingRequest),
}

type wsIncomingRequest interface {
	Validate(ctx context.Context) app.Status
	Execute(ctx context.Context) (wsIncomingReply, app.Status)
}

type wsIncomingReply interface{}
