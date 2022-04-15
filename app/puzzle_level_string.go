// Code generated by "stringer -type=PuzzleLevel -linecomment -trimprefix PuzzleLevel -output puzzle_level_string.go"; DO NOT EDIT.

package app

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PuzzleLevelEasy-1]
	_ = x[PuzzleLevelNormal-2]
	_ = x[PuzzleLevelHard-3]
	_ = x[PuzzleLevelHarder-4]
	_ = x[PuzzleLevelInsane-5]
	_ = x[PuzzleLevelDemon-6]
	_ = x[PuzzleLevelCustom-7]
}

const _PuzzleLevel_name = "EasyNormalHardHarderInsaneDemonCustom"

var _PuzzleLevel_index = [...]uint8{0, 4, 10, 14, 20, 26, 31, 37}

func (i PuzzleLevel) String() string {
	i -= 1
	if i >= PuzzleLevel(len(_PuzzleLevel_index)-1) {
		return "PuzzleLevel(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _PuzzleLevel_name[_PuzzleLevel_index[i]:_PuzzleLevel_index[i+1]]
}
