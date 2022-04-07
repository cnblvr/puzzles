package sudoku_classic

import (
	crand "crypto/rand"
	"encoding/binary"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	"math/rand"
)

const size = 9

type puzzle [size][size]uint8

func Parse(s string) (app.PuzzleAssistant, error) {
	if len(s) != size*size {
		return nil, errors.Errorf("invalid puzzle length: %d", len(s))
	}
	var p puzzle
	for i := 0; i < size*size; i++ {
		if '1' <= s[i] && s[i] <= '9' {
			p[i/size][i%size] = s[i] - '0'
		}
	}
	return &p, nil
}

func (p puzzle) String() string {
	out := make([]byte, 81)
	for i := 0; i < size*size; i++ {
		char := p[i/size][i%size]
		if char > 0 {
			out[i] = char + '0'
		} else {
			out[i] = '.'
		}
	}
	return string(out)
}

func NewRandomSolution() (s app.PuzzleAssistant, seed int64) {
	seedBts := make([]byte, 8)
	if _, err := crand.Reader.Read(seedBts); err != nil {
		panic(err)
	}
	seed = int64(binary.LittleEndian.Uint64(seedBts))
	return NewSolutionBySeed(seed), seed
}

func NewSolutionBySeed(seed int64) app.PuzzleAssistant {
	rnd := rand.New(rand.NewSource(seed))

	s := generateWithoutShuffling(rnd)

	return &s
}

func generateWithoutShuffling(rnd *rand.Rand) (s puzzle) {
	digits := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9}
	i := 0
	for len(digits) > 0 {
		idx := rnd.Int() % len(digits)
		s[0][i] = digits[idx]
		i++
		digits = append(digits[:idx], digits[idx+1:]...)
	}

	for l := 1; l < size; l++ {
		if l%3 == 0 {
			copy(s[l][:size-1], s[l-1][1:size])
			s[l][size-1] = s[l-1][0]
			continue
		}
		copy(s[l][:size-3], s[l-1][3:size])
		copy(s[l][size-3:size], s[l-1][:3])
	}

	return
}

func (p *puzzle) Rotate(r app.RotationType) {
	r = r % 4

	switch r {
	case app.RotateTo90:
		// reflect along the main diagonal
		// 1234 -> 1342
		// 3412    2413
		// 4123    3124
		// 2341    4231
		p.reflectAlongMainDiagonal()
		// reflect vertically
		// 1342 -> 4231
		// 2413    3124
		// 3124    2413
		// 4231    1342
		p.Reflect(app.Vertical)

	case app.RotateTo180:
		// reflect vertically
		// 1234    2341
		// 3412    4123
		// 4123    3412
		// 2341    1234
		p.Reflect(app.Vertical)
		// reflect horizontally
		// 2341    1432
		// 4123    3214
		// 3412    2143
		// 1234    4321
		p.Reflect(app.Horizontal)

	case app.RotateTo270:
		// reflect along the main diagonal
		// 1234 -> 1342
		// 3412    2413
		// 4123    3124
		// 2341    4231
		p.reflectAlongMainDiagonal()
		// 1342 -> 2431
		// 2413    3142
		// 3124    4213
		// 4231    1324
		p.Reflect(app.Horizontal)
	}
}

func (p *puzzle) Reflect(d app.DirectionType) {
	d = d % 2
	switch d {
	case app.Horizontal:
		for row := 0; row < size; row++ {
			for i := 0; i < size/2; i++ {
				p[row][i], p[row][size-1-i] = p[row][size-1-i], p[row][i]
			}
		}
	case app.Vertical:
		for col := 0; col < size; col++ {
			for i := 0; i < size/2; i++ {
				p[i][col], p[size-1-i][col] = p[size-1-i][col], p[i][col]
			}
		}
	}
}

func (p *puzzle) reflectAlongMainDiagonal() {
	for diag := 0; diag < size; diag++ {
		for i := diag + 1; i < size; i++ {
			p[diag][i], p[i][diag] = p[i][diag], p[diag][i]
		}
	}
}
