package frontend

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
)

type websocketMessage struct {
	Method string          `json:"method"`
	Echo   string          `json:"echo,omitempty"`
	Error  string          `json:"error,omitempty"`
	Body   json.RawMessage `json:"body,omitempty"`
}

func (srv *service) HandleGameWs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log, session := FromContextLogger(ctx), FromContextSession(ctx)
	log = log.With().Int64("session_id", session.SessionID).Logger()
	if session.UserID > 0 {
		log = log.With().Int64("user_id", session.UserID).Logger()
	} else {
		log = log.With().Bool("anonymous", true).Logger()
	}
	conn, err := srv.gameWebsocket.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade client")
		return
	}
	defer conn.Close()
	for {
		ctx := context.Background()
		ctx = NewContextLogger(ctx, log)
		ctx = NewContextSession(ctx, session)
		ctx = NewContextServiceFrontend(ctx, srv)
		var req websocketMessage
		mType, reqBts, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
			) {
				log.Debug().Err(err).Msg("connection closed")
				return
			}
			log.Error().Err(err).Msg("failed to read message")
			return
		}
		if mType != websocket.TextMessage {
			log.Debug().Msg("message is not TextMessage")
			continue
		}
		log.Debug().Msgf("ws request:  %s", reqBts)
		if err := json.Unmarshal(reqBts, &req); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal request")
			return
		}
		resp := websocketMessage{
			Method: req.Method,
			Echo:   req.Echo,
		}
		resp.Body, resp.Error = websocketRequestExecute(ctx, req.Method, req.Body)
		respBts, err := json.Marshal(resp)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal response")
			return
		}
		log.Debug().Msgf("ws response: %s", respBts)
		if err := conn.WriteMessage(websocket.TextMessage, respBts); err != nil {
			log.Error().Err(err).Msg("failed to write message")
			return
		}
	}
}

func websocketRequestExecute(ctx context.Context, method string, reqBody []byte) ([]byte, string) {
	reqObj, err := websocketPool.GetRequest(method)
	if err != nil {
		log.Warn().Err(err).Msg("failed to find request")
		return nil, err.Error()
	}
	if len(reqBody) > 0 {
		if err := json.Unmarshal(reqBody, reqObj); err != nil {
			log.Warn().Err(err).Msg("failed to unmarshal body request")
			return nil, "body invalid"
		}
	}
	if err := reqObj.Validate(ctx); err != nil {
		log.Error().Err(err).Msg("failed to validate")
		return nil, err.Error()
	}
	respObj, err := reqObj.Execute(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute")
		return nil, err.Error()
	}
	respBody, err := json.Marshal(respObj)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal body response")
		return nil, "internal server error"
	}
	return respBody, ""
}
