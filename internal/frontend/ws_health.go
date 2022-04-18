package frontend

import (
	"context"
	"github.com/cnblvr/puzzles/app"
)

func init() {
	wsAddIncoming("health", (*wsHealthRequest)(nil))
}

type wsHealthRequest string

func (r *wsHealthRequest) Validate(ctx context.Context) app.Status {
	return nil
}

func (r *wsHealthRequest) Execute(ctx context.Context) (wsIncomingReply, app.Status) {
	resp := wsHealthReply("OK")
	return &resp, nil
}

// TODO handle and test
type wsHealthReply string
