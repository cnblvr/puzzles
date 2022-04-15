package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func init() {
	websocketPool.Add((*websocketGetPuzzleRequest)(nil), (*websocketGetPuzzleResponse)(nil))
}

type websocketGetPuzzleRequest struct {
	GameID uuid.UUID `json:"game_id"`
}

func (websocketGetPuzzleRequest) Method() string {
	return "getPuzzle"
}

func (r websocketGetPuzzleRequest) Validate(ctx context.Context) error {
	if r.GameID == uuid.Nil {
		return errors.Errorf("game_id is empty")
	}
	return nil
}

func (r websocketGetPuzzleRequest) Execute(ctx context.Context) (websocketResponse, error) {
	srv := FromContextServiceFrontendOrNil(ctx)

	puzzle, game, err := srv.puzzleRepository.GetPuzzleAndGame(ctx, r.GameID)
	if err != nil {
		return websocketGetPuzzleResponse{}, fmt.Errorf("internal server error")
	}

	resp := websocketGetPuzzleResponse{
		Puzzle:     puzzle.Clues,
		Candidates: json.RawMessage("{}"),
		IsNew:      game.IsNew,
		IsWin:      game.IsWin,
	}
	if game.CandidatesAtStart {
		resp.Candidates = json.RawMessage(puzzle.Candidates)
	}

	statePuzzle, err := puzzle_library.GetAssistant(puzzle.Type, resp.StatePuzzle)
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}

	if game.IsNew {
		resp.StateCandidates = resp.Candidates
		game.State = puzzle.Clues
		game.StateCandidates = string(resp.StateCandidates)
		if err := srv.puzzleRepository.UpdatePuzzleGame(ctx, game); err != nil {
			return websocketGetPuzzleResponse{}, fmt.Errorf("internal server error")
		}
	} else {
		resp.StatePuzzle = game.State
		resp.StateCandidates = json.RawMessage(game.StateCandidates)
		resp.Errors = statePuzzle.GetWrongPoints()
		wrongCandidates, err := statePuzzle.GetWrongCandidates(game.StateCandidates)
		if err != nil {
			return websocketGetPuzzleResponse{}, fmt.Errorf("bad request")
		}
		resp.ErrorsCandidates = json.RawMessage(wrongCandidates)
	}

	return resp, nil
}

// TODO handle and test
type websocketGetPuzzleResponse struct {
	Puzzle     string          `json:"puzzle"`
	Candidates json.RawMessage `json:"candidates,omitempty"`
	IsNew      bool            `json:"is_new,omitempty"`
	IsWin      bool            `json:"is_win,omitempty"`

	// if IsNew is false
	StatePuzzle      string          `json:"state_puzzle,omitempty"`
	StateCandidates  json.RawMessage `json:"state_candidates,omitempty"`
	Errors           []app.Point     `json:"errors,omitempty"`
	ErrorsCandidates json.RawMessage `json:"errorsCandidates,omitempty"`
}

func (websocketGetPuzzleResponse) Method() string {
	return "getPuzzle"
}

func (r websocketGetPuzzleResponse) Validate(ctx context.Context) error {
	return nil
}

func (r websocketGetPuzzleResponse) Execute(ctx context.Context) error {
	return nil
}
