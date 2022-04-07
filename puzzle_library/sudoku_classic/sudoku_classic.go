package sudoku_classic

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"strings"
)

const (
	// size is the width and height measurement
	size = 9
	// sizeGrp is the number of rows or columns for the big line
	sizeGrp = 3
)

type puzzle [size][size]uint8

func parse(str string) (*puzzle, error) {
	if len(str) != size*size {
		return nil, errors.Errorf("invalid puzzle length: %d", len(str))
	}
	var p puzzle
	for i := 0; i < size*size; i++ {
		if '1' <= str[i] && str[i] <= '9' {
			p[i/size][i%size] = str[i] - '0'
		}
	}
	return &p, nil
}

// ParseAssistant parses str into an interface that can be used to work with the
// generated puzzle or user state of the puzzle.
func ParseAssistant(str string) (app.PuzzleAssistant, error) {
	return parse(str)
}

// ParseGenerator parses str into an interface that can be used to generate the
// puzzle.
func ParseGenerator(str string) (app.PuzzleGenerator, error) {
	return parse(str)
}

func (p puzzle) String() string {
	out := make([]byte, size*size)
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

// NewRandomSolution generates a solution randomly for further extraction of
// digits.
func NewRandomSolution() (s app.PuzzleGenerator, seed int64) {
	seedBts := make([]byte, 8)
	if _, err := crand.Reader.Read(seedBts); err != nil {
		panic(err)
	}
	seed = int64(binary.LittleEndian.Uint64(seedBts))
	return NewSolutionBySeed(seed), seed
}

// NewSolutionBySeed generates a solution with a given seed for further
// extraction of digits.
func NewSolutionBySeed(seed int64) app.PuzzleGenerator {
	rnd := rand.New(rand.NewSource(seed))

	s := generateWithoutShuffling(rnd)

	return &s
}

func generateWithoutShuffling(rnd *rand.Rand) (s puzzle) {
	digits := []uint8{1, 2, 3, 4, 5, 6, 7, 8, 9}
	// Generate first line randomly
	i := 0
	for len(digits) > 0 {
		idx := rnd.Int() % len(digits)
		s[0][i] = digits[idx]
		i++
		digits = append(digits[:idx], digits[idx+1:]...)
	}

	// The second and third lines is the offset of the previous line to the left by 3
	// The next "big" lines (d-f and g-i) are generated like this:
	//  lines d and g:    offset of the previous line to the left by 1
	//  lines e, f, h, i: offset of the previous line to the left by 3
	for l := 1; l < size; l++ {
		if l%sizeGrp == 0 {
			copy(s[l][:size-1], s[l-1][1:size])
			s[l][size-1] = s[l-1][0]
			continue
		}
		copy(s[l][:size-sizeGrp], s[l-1][sizeGrp:size])
		copy(s[l][size-sizeGrp:size], s[l-1][:sizeGrp])
	}

	return
}

func (p *puzzle) SwapLines(dir app.DirectionType, a, b int) error {
	switch dir {
	case app.Horizontal, app.Vertical:
	default:
		return errors.Errorf("dir unknown")
	}
	if 0 > a || a > size-1 {
		return errors.Errorf("a is incorrect line")
	}
	if 0 > b || b > size-1 {
		return errors.Errorf("b is incorrect line")
	}
	if a == b {
		return errors.Errorf("a == b")
	}
	for i := 0; i < size; i++ {
		if dir == app.Horizontal {
			p[a][i], p[b][i] = p[b][i], p[a][i]
		} else {
			p[i][a], p[i][b] = p[i][b], p[i][a]
		}
	}
	return nil
}

func (p *puzzle) SwapBigLines(dir app.DirectionType, a, b int) error {
	switch dir {
	case app.Horizontal, app.Vertical:
	default:
		return errors.Errorf("dir unknown")
	}
	if 0 > a || a > sizeGrp-1 {
		return errors.Errorf("a is incorrect line")
	}
	if 0 > b || b > sizeGrp-1 {
		return errors.Errorf("b is incorrect line")
	}
	if a == b {
		return errors.Errorf("a == b")
	}
	for l := 0; l < sizeGrp; l++ {
		for i := 0; i < size; i++ {
			la, lb := a*sizeGrp+l, b*sizeGrp+l
			if dir == app.Horizontal {
				p[la][i], p[lb][i] = p[lb][i], p[la][i]
			} else {
				p[i][la], p[i][lb] = p[i][lb], p[i][la]
			}
		}
	}
	return nil
}

func (p *puzzle) Rotate(r app.RotationType) error {
	// A 0-degree angle rotation has no effect.
	// The input argument r is the ring [0; 3].
	r = r % (app.RotateTo270 + 1)

	// The rotation of the board of this type of Sudoku to any angle is done by vertical, horizontal reflections and
	// reflection along the major diagonal. So this method is useless.

	var err1, err2 error
	switch r {
	case app.RotateTo90:
		err1 = p.Reflect(app.ReflectMajorDiagonal)
		err2 = p.Reflect(app.ReflectVertical)

	case app.RotateTo180:
		err1 = p.Reflect(app.ReflectVertical)
		err2 = p.Reflect(app.ReflectHorizontal)

	case app.RotateTo270:
		err1 = p.Reflect(app.ReflectMajorDiagonal)
		err2 = p.Reflect(app.ReflectHorizontal)
	}
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (p *puzzle) Reflect(r app.ReflectionType) error {
	switch r {
	case app.ReflectHorizontal:
		for row := 0; row < size; row++ {
			for i := 0; i < size/2; i++ {
				p[row][i], p[row][size-1-i] = p[row][size-1-i], p[row][i]
			}
		}

	case app.ReflectVertical:
		for col := 0; col < size; col++ {
			for i := 0; i < size/2; i++ {
				p[i][col], p[size-1-i][col] = p[size-1-i][col], p[i][col]
			}
		}

	case app.ReflectMajorDiagonal:
		for diag := 0; diag < size; diag++ {
			for i := diag + 1; i < size; i++ {
				p[diag][i], p[i][diag] = p[i][diag], p[diag][i]
			}
		}

	case app.ReflectMinorDiagonal:
		for diag := 0; diag < size; diag++ {
			for i := 0; i < size-1-diag; i++ {
				p[diag][i], p[size-1-i][size-1-diag] = p[size-1-i][size-1-diag], p[diag][i]
			}
		}
	default:
		return errors.Errorf("unknown reflection type %d", r)
	}
	return nil
}

func (p *puzzle) SwapDigits(a, b uint8) error {
	if 1 > a || a > size {
		return errors.Errorf("a is incorrect digit")
	}
	if 1 > b || b > size {
		return errors.Errorf("b is incorrect digit")
	}
	if a == b {
		return errors.Errorf("a == b")
	}
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			switch p[row][col] {
			case a:
				p[row][col] = b
			case b:
				p[row][col] = a
			}
		}
	}
	return nil
}

// ASCII representation of the puzzle when debugging.
func (p puzzle) debug() string {
	var out strings.Builder
	out.WriteString("╔═══════╤═══════╤═══════╗\n")
	for i := 0; i < size; i++ {
		out.WriteString("║ ")
		for j := 0; j < size; j++ {
			if val := int(p[i][j]); val == 0 {
				out.WriteByte(' ')
			} else {
				out.WriteString(strconv.Itoa(val))
			}
			if j%sizeGrp == sizeGrp-1 && j != size-1 {
				out.WriteString(" │ ")
			} else {
				out.WriteByte(' ')
			}
		}
		out.WriteString(fmt.Sprintf("║ %s\n", string('a'+byte(i))))
		if i%sizeGrp == sizeGrp-1 && i != size-1 {
			out.WriteString("╟───────┼───────┼───────╢\n")
		}
	}
	out.WriteString("╚═══════╧═══════╧═══════╝\n")
	out.WriteString("  1 2 3   4 5 6   7 8 9  ")
	return out.String()
}
