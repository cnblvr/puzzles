package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cnblvr/puzzles/puzzle_library"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func init() {
	websocketPool.Add((*websocketGetPuzzleRequest)(nil), (*websocketGetPuzzleResponse)(nil))
}

type websocketGetPuzzleRequest struct {
	GameID         uuid.UUID `json:"game_id"`
	NeedCandidates bool      `json:"need_candidates,omitempty"`
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
		Puzzle: puzzle.Clues,
	}

	assistant, err := puzzle_library.GetAssistant(puzzle.Type)
	if err != nil {
		return websocketGetPuzzleResponse{}, fmt.Errorf("internal server error")
	}
	if r.NeedCandidates {
		game.StateCandidates = assistant.GetCandidates(ctx, puzzle.Clues)
		resp.Candidates = json.RawMessage(game.StateCandidates)
		if err := srv.puzzleRepository.UpdatePuzzleGame(ctx, game); err != nil {
			return websocketGetPuzzleResponse{}, fmt.Errorf("internal server error")
		}
	}

	return resp, nil
}

// TODO handle and test
type websocketGetPuzzleResponse struct {
	Puzzle     string          `json:"puzzle"`
	Candidates json.RawMessage `json:"candidates,omitempty"`
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
