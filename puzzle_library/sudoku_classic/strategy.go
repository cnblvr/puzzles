package sudoku_classic

import (
	"fmt"
	"github.com/cnblvr/puzzles/app"
)

type puzzleStep interface {
	Strategy() string
	Description() string
}

type puzzleStepSet struct {
	strategy           string
	point              app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepSet) Strategy() string {
	return s.strategy
}

func (s puzzleStepSet) Description() string {
	out := fmt.Sprintf("set %d in point %s", s.value, s.point)
	if len(s.removalsCandidates) > 0 {
		out += fmt.Sprintf(" with removals in %v", s.removalsCandidates)
	}
	return out
}

type puzzleStepNakedPairOrTriple struct {
	points             []app.Point
	set                []uint8
	removalsCandidates []app.Point
}

func (s puzzleStepNakedPairOrTriple) Strategy() string {
	if len(s.points) == 2 {
		return "Naked Pair"
	} else {
		return "Naked Triple"
	}
}

func (s puzzleStepNakedPairOrTriple) Description() string {
	return fmt.Sprintf("has candidates %v in points %s and remove candidates in points %v", s.set, s.points, s.removalsCandidates)
}

type puzzleStepHiddenPairOrTriple struct {
	points []app.Point
	set    []uint8
}

func (s puzzleStepHiddenPairOrTriple) Strategy() string {
	if len(s.set) == 2 {
		return "Hidden Pair"
	} else {
		return "Hidden Triple"
	}
}

func (s puzzleStepHiddenPairOrTriple) Description() string {
	return fmt.Sprintf("has candidates %v in points %s", s.set, s.points)
}

type puzzleStepPointingPairOrTriple struct {
	points             []app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepPointingPairOrTriple) Strategy() string {
	if len(s.points) == 2 {
		return "Pointing Pair"
	} else {
		return "Pointing Triple"
	}
}

func (s puzzleStepPointingPairOrTriple) Description() string {
	return fmt.Sprintf("has candidate %d in points %v and remove candidates in points %v", s.value, s.points, s.removalsCandidates)
}

type puzzleStepBoxLineReductionPairOrTriple struct {
	points             []app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepBoxLineReductionPairOrTriple) Strategy() string {
	if len(s.points) == 2 {
		return "Box/Line Reduction Pair"
	} else {
		return "Box/Line Reduction Triple"
	}
}

func (s puzzleStepBoxLineReductionPairOrTriple) Description() string {
	return fmt.Sprintf("has candidate %d in points %v and remove candidates in points %v", s.value, s.points, s.removalsCandidates)
}

type puzzleStepXWing struct {
	pairA              []app.Point
	pairB              []app.Point
	value              uint8
	removalsCandidates []app.Point
}

func (s puzzleStepXWing) Strategy() string {
	return "X-Wing"
}

func (s puzzleStepXWing) Description() string {
	return fmt.Sprintf("has candidate %d in pairs %v and %v; remove candidates in points %v", s.value, s.pairA, s.pairB, s.removalsCandidates)
}
