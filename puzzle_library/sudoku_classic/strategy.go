package sudoku_classic

import (
	"fmt"
	"github.com/cnblvr/puzzles/app"
)

type puzzleStepSetter interface {
	app.PuzzleStep
	setCandidateChanges(string)
}

type candidateChanges struct {
	changes string
}

func (c *candidateChanges) setCandidateChanges(s string) {
	c.changes = s
}

func (c candidateChanges) CandidateChanges() string {
	return c.changes
}

type puzzleStepSet struct {
	candidateChanges
	strategy app.PuzzleStrategy
	point    app.Point
	value    uint8
}

func (s puzzleStepSet) Strategy() app.PuzzleStrategy {
	return s.strategy
}

func (s puzzleStepSet) Description() string {
	out := fmt.Sprintf("set %d in point %s", s.value, s.point)
	return out
}

type puzzleStepNakedStrategy struct {
	candidateChanges
	points []app.Point
	set    []uint8
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
	return fmt.Sprintf("has candidates %v in points %s", s.set, s.points)
}

type puzzleStepHiddenStrategy struct {
	candidateChanges
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
	candidateChanges
	points []app.Point
	value  uint8
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
	return fmt.Sprintf("has candidate %d in points %v", s.value, s.points)
}

type puzzleStepBoxLineReductionStrategy struct {
	candidateChanges
	points []app.Point
	value  uint8
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
	return fmt.Sprintf("has candidate %d in points %v", s.value, s.points)
}

type puzzleStepXWingStrategy struct {
	candidateChanges
	pairA []app.Point
	pairB []app.Point
	value uint8
}

func (s puzzleStepXWingStrategy) Strategy() app.PuzzleStrategy {
	return app.StrategyXWing
}

func (s puzzleStepXWingStrategy) Description() string {
	return fmt.Sprintf("has candidate %d in pairs %v and %v", s.value, s.pairA, s.pairB)
}
