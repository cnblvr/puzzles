package frontend

import (
	"context"
	"encoding/json"
	"github.com/cnblvr/puzzles/app"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net/http"
)

type websocketMessage struct {
	Method    string          `json:"method"`
	Echo      string          `json:"echo,omitempty"`
	Error     string          `json:"error,omitempty"`
	ErrorCode uint16          `json:"errorCode,omitempty"`
	Body      json.RawMessage `json:"body,omitempty"`
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
		resp.Body, err = websocketRequestExecute(ctx, req.Method, req.Body)
		if err != nil {
			if status, ok := err.(app.Status); ok {
				resp.Error, resp.ErrorCode = status.GetMessage(), status.GetCode()
			} else {
				resp.Error, resp.ErrorCode = status.Error(), app.StatusUnknown.GetCode()
			}
		}
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

func websocketRequestExecute(ctx context.Context, method string, reqBody []byte) ([]byte, error) {
	reqObj, err := wsGetIncoming(method)
	if err != nil {
		return nil, app.StatusMethodNotAllowed.WithError(errors.WithStack(err))
	}
	if len(reqBody) > 0 {
		if err := json.Unmarshal(reqBody, reqObj); err != nil {
			return nil, app.StatusBadRequest.WithError(errors.Wrapf(err, "failed to decode request"))
		}
	}
	if mw, ok := reqObj.(wsGameMiddlewareInterface); ok {
		if status := mw.GameMiddleware(ctx); status != nil {
			return nil, status
		}
	}
	if status := reqObj.Validate(ctx); status != nil {
		return nil, status
	}
	respObj, status := reqObj.Execute(ctx)
	if err != nil {
		return nil, status
	}
	respBody, err := json.Marshal(respObj)
	if err != nil {
		return nil, app.StatusInternalServerError.WithError(errors.WithStack(err))
	}
	return respBody, nil
}

type wsGameMiddleware struct {
	GameID uuid.UUID `json:"game_id"`
	puzzle *app.Puzzle
	game   *app.PuzzleGame
}

type wsGameMiddlewareInterface interface {
	GameMiddleware(ctx context.Context) app.Status
}

func (m *wsGameMiddleware) GameMiddleware(ctx context.Context) app.Status {
	srv := FromContextServiceFrontendOrNil(ctx)
	if m.GameID == uuid.Nil {
		return app.StatusBadRequest.WithMessage("game_id is empty")
	}

	var err error
	m.puzzle, m.game, err = srv.puzzleRepository.GetPuzzleAndGame(ctx, m.GameID)
	switch {
	case err == nil:
	case errors.Is(err, app.ErrorPuzzleGameNotFound):
		return app.StatusBadRequest.WithMessage("game not found").WithError(errors.WithStack(err))
	case errors.Is(err, app.ErrorPuzzleNotFound):
		return app.StatusBadRequest.WithMessage("game not found").WithError(errors.WithStack(err))
	default:
		return app.StatusInternalServerError.WithError(errors.WithStack(err))
	}

	return nil
}
