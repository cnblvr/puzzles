package app

const (
	DefaultPuzzleType = PuzzleSudokuClassic

	DefaultPuzzleLevel = PuzzleLevelNormal

	DefaultCandidatesAtStart = false

	DefaultUseHighlights = false

	DefaultShowCandidates = true

	DefaultShowWrongs = false
)

func (up *UserPreferences) Defaults() {
	up.PuzzleType = DefaultPuzzleType
	up.PuzzleLevel = DefaultPuzzleLevel
	up.CandidatesAtStart = DefaultCandidatesAtStart
	up.UseHighlights = DefaultUseHighlights
	up.ShowCandidates = DefaultShowCandidates
	up.ShowWrongs = DefaultShowWrongs
}
