package frontend

import (
	"context"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/cnblvr/puzzles/puzzle_library"
	"github.com/google/uuid"
	"sort"
)

func init() {
	websocketPool.Add((*websocketMakeStepRequest)(nil), (*websocketMakeStepResponse)(nil))
}

type websocketMakeStepRequest struct {
	GameID         uuid.UUID `json:"game_id"`
	State          string    `json:"state"`
	NeedCandidates bool      `json:"need_candidates,omitempty"`
}

func (websocketMakeStepRequest) Method() string {
	return "makeStep"
}

func (r websocketMakeStepRequest) Validate(ctx context.Context) error {
	if r.GameID == uuid.Nil {
		return fmt.Errorf("game_id is empty")
	}
	//if len(r.State) < 81 {
	//	return fmt.Errorf("state format invalid")
	//}
	return nil
}

func (r websocketMakeStepRequest) Execute(ctx context.Context) (websocketResponse, error) {
	srv := FromContextServiceFrontendOrNil(ctx)

	uniqueErrs := make(map[app.Point]struct{})

	puzzle, err := srv.puzzleRepository.GetPuzzleByGameID(ctx, r.GameID)
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}

	if r.State == puzzle.Solution {
		// WIN
		return websocketMakeStepResponse{
			Win: true,
		}, nil
	}

	assistant, err := puzzle_library.GetAssistant(puzzle.Type)
	if err != nil {
		return websocketMakeStepResponse{}, fmt.Errorf("internal server error")
	}
	for _, p := range assistant.FindUserErrors(ctx, r.State) {
		uniqueErrs[p] = struct{}{}
	}

	// TODO new method "compare with answer" and use this function
	//board := sudoku_classic.PuzzleFromString(boardStr)
	//for _, p := range board.FindErrors(userState) {
	//	uniqueErrs[p] = struct{}{}
	//}

	resp := websocketMakeStepResponse{}
	for p := range uniqueErrs {
		resp.Errors = append(resp.Errors, p)
	}
	sort.Slice(resp.Errors, func(i, j int) bool {
		if resp.Errors[i].Row != resp.Errors[j].Row {
			return resp.Errors[i].Row < resp.Errors[j].Row
		}
		return resp.Errors[i].Col < resp.Errors[j].Col
	})
	if r.NeedCandidates {
		resp.Candidates = assistant.GetCandidates(ctx, r.State)
	}
	return resp, nil
}

// TODO handle and test
type websocketMakeStepResponse struct {
	Errors     []app.Point          `json:"errors,omitempty"`
	Win        bool                 `json:"win,omitempty"`
	Candidates app.PuzzleCandidates `json:"candidates,omitempty"`
}

func (websocketMakeStepResponse) Method() string {
	return "makeStep"
}

func (r websocketMakeStepResponse) Validate(ctx context.Context) error {
	return nil
}

func (r websocketMakeStepResponse) Execute(ctx context.Context) error {
	return nil
}
