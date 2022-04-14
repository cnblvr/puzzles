package sudoku_classic

import (
	"fmt"
	"github.com/cnblvr/puzzles/app"
)

type puzzleStepSet struct {
	strategy           app.PuzzleStrategy
	point              app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepSet) Strategy() app.PuzzleStrategy {
	return s.strategy
}

func (s puzzleStepSet) Description() string {
	out := fmt.Sprintf("set %d in point %s", s.value, s.point)
	if len(s.removalsCandidates) > 0 {
		out += fmt.Sprintf(" with removals in %v", s.removalsCandidates)
	}
	return out
}

type puzzleStepNakedStrategy struct {
	points             []app.Point
	set                []uint8
	removalsCandidates []app.Point
}

func (s puzzleStepNakedStrategy) Strategy() app.PuzzleStrategy {
	switch len(s.points) {
	case 2:
		return app.StrategyNakedPair
	case 3:
		return app.StrategyNakedTriple
	case 4:
		return app.StrategyNakedQuad
	default:
		return app.StrategyUnknown
	}
}

func (s puzzleStepNakedStrategy) Description() string {
	return fmt.Sprintf("has candidates %v in points %s and remove candidates in points %v", s.set, s.points, s.removalsCandidates)
}

type puzzleStepHiddenStrategy struct {
	points []app.Point
	set    []uint8
}

func (s puzzleStepHiddenStrategy) Strategy() app.PuzzleStrategy {
	switch len(s.set) {
	case 2:
		return app.StrategyHiddenPair
	case 3:
		return app.StrategyHiddenTriple
	case 4:
		return app.StrategyHiddenQuad
	default:
		return app.StrategyUnknown
	}
}

func (s puzzleStepHiddenStrategy) Description() string {
	return fmt.Sprintf("has candidates %v in points %s", s.set, s.points)
}

type puzzleStepPointingStrategy struct {
	points             []app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepPointingStrategy) Strategy() app.PuzzleStrategy {
	switch len(s.points) {
	case 2:
		return app.StrategyPointingPair
	case 3:
		return app.StrategyPointingTriple
	default:
		return app.StrategyUnknown
	}
}

func (s puzzleStepPointingStrategy) Description() string {
	return fmt.Sprintf("has candidate %d in points %v and remove candidates in points %v", s.value, s.points, s.removalsCandidates)
}

type puzzleStepBoxLineReductionStrategy struct {
	points             []app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepBoxLineReductionStrategy) Strategy() app.PuzzleStrategy {
	switch len(s.points) {
	case 2:
		return app.StrategyBoxLineReductionPair
	case 3:
		return app.StrategyBoxLineReductionTriple
	default:
		return app.StrategyUnknown
	}
}

func (s puzzleStepBoxLineReductionStrategy) Description() string {
	return fmt.Sprintf("has candidate %d in points %v and remove candidates in points %v", s.value, s.points, s.removalsCandidates)
}

type puzzleStepXWingStrategy struct {
	pairA              []app.Point
	pairB              []app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepXWingStrategy) Strategy() app.PuzzleStrategy {
	return app.StrategyXWing
}

func (s puzzleStepXWingStrategy) Description() string {
	return fmt.Sprintf("has candidate %d in pairs %v and %v; remove candidates in points %v", s.value, s.pairA, s.pairB, s.removalsCandidates)
}
