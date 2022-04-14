// Code generated by "stringer -type=PuzzleStrategy -linecomment -trimprefix Strategy -output puzzle_strategy_string.go"; DO NOT EDIT.

package app

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[StrategyNakedSingle-1]
	_ = x[StrategyNakedPair-2]
	_ = x[StrategyNakedTriple-4]
	_ = x[StrategyNakedQuad-8]
	_ = x[StrategyHiddenSingle-16]
	_ = x[StrategyHiddenPair-32]
	_ = x[StrategyHiddenTriple-64]
	_ = x[StrategyHiddenQuad-128]
	_ = x[StrategyPointingPair-256]
	_ = x[StrategyPointingTriple-512]
	_ = x[StrategyBoxLineReductionPair-1024]
	_ = x[StrategyBoxLineReductionTriple-2048]
	_ = x[StrategyXWing-4096]
	_ = x[StrategyUnknown-0]
}

const _PuzzleStrategy_name = "UnknownNaked SingleNaked PairNaked TripleNaked QuadHidden SingleHidden PairHidden TripleHidden QuadPointing PairPointing TripleBox/Line Reduction PairBox/Line Reduction TripleX-Wing"

var _PuzzleStrategy_map = map[PuzzleStrategy]string{
	0:    _PuzzleStrategy_name[0:7],
	1:    _PuzzleStrategy_name[7:19],
	2:    _PuzzleStrategy_name[19:29],
	4:    _PuzzleStrategy_name[29:41],
	8:    _PuzzleStrategy_name[41:51],
	16:   _PuzzleStrategy_name[51:64],
	32:   _PuzzleStrategy_name[64:75],
	64:   _PuzzleStrategy_name[75:88],
	128:  _PuzzleStrategy_name[88:99],
	256:  _PuzzleStrategy_name[99:112],
	512:  _PuzzleStrategy_name[112:127],
	1024: _PuzzleStrategy_name[127:150],
	2048: _PuzzleStrategy_name[150:175],
	4096: _PuzzleStrategy_name[175:181],
}

func (i PuzzleStrategy) String() string {
	if str, ok := _PuzzleStrategy_map[i]; ok {
		return str
	}
	return "PuzzleStrategy(" + strconv.FormatInt(int64(i), 10) + ")"
}
