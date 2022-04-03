package frontend

import (
	"context"
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

	puzzle, err := srv.puzzleRepository.GetPuzzleByGameID(ctx, r.GameID)
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
		resp.Candidates = assistant.GetCandidates(ctx, puzzle.Clues)
	}

	return resp, nil
}

// TODO handle and test
type websocketGetPuzzleResponse struct {
	Puzzle     string               `json:"puzzle"`
	Candidates app.PuzzleCandidates `json:"candidates,omitempty"`
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
