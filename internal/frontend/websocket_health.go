package frontend

import (
	"context"
)

func init() {
	websocketPool.Add((*websocketHealthRequest)(nil), (*websocketHealthResponse)(nil))
}

type websocketHealthRequest string

func (websocketHealthRequest) Method() string {
	return "health"
}

func (r websocketHealthRequest) Validate(ctx context.Context) error {
	return nil
}

func (r websocketHealthRequest) Execute(ctx context.Context) (websocketResponse, error) {
	return websocketHealthResponse("OK"), nil
}

// TODO handle and test
type websocketHealthResponse string

func (websocketHealthResponse) Method() string {
	return "health"
}

func (r websocketHealthResponse) Validate(ctx context.Context) error {
	return nil
}

func (r websocketHealthResponse) Execute(ctx context.Context) error {
	return nil
}
