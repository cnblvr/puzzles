package sudoku_classic

import (
	"encoding/json"
	"fmt"
	"github.com/cnblvr/puzzles/app"
	"math/rand"
	"strconv"
	"testing"
)

func TestDebug(t *testing.T) {
	p, err := parse("....41....6....2...........32.6.........5..417...........2..3...48......5.1......")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("debug clues\n%s", p.debug())
	t.Logf("debug candidates\n%s", p.findSimpleCandidates().debug(nil))
	t.Logf("debug candidates and clues\n%s", p.findSimpleCandidates().debug(p))
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantOut string
		wantErr bool
	}{
		{
			name:    "empty",
			in:      "",
			wantErr: true,
		},
		{
			name:    "few bytes",
			in:      "123456789",
			wantErr: true,
		},
		{
			name:    "many bytes",
			in:      "123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789",
			wantErr: true,
		},
		{
			name:    "success",
			in:      "123456789123456789123456789123456789123456789123456789123456789123456789123456789",
			wantOut: "123456789123456789123456789123456789123456789123456789123456789123456789123456789",
		},
		{
			name:    "no clues",
			in:      ".................................................................................",
			wantOut: ".................................................................................",
		},
		{
			name:    "no clues: any characters",
			in:      `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ,.[];'/!@#$%^&*()_+-="{}:?\ ` + "`",
			wantOut: ".................................................................................",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParseAssistant(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("ParseAssistant() error = %v, wantErr is false", err)
					return
				}
				return
			} else {
				if tt.wantErr {
					t.Errorf("ParseAssistant() error = <nil>, wantErr is true")
					return
				}
			}
			out := p.String()
			if out != tt.wantOut {
				t.Errorf("ParseAssistant().String()\ngot  = %s\nwant = %s", out, tt.wantOut)
				return
			}
		})
	}
}

func Test_generateWithoutShuffling(t *testing.T) {
	tests := []struct {
		name string
		seed int64
		want string
	}{
		{
			name: "seed 0",
			seed: 0,
			want: "316579248579248316248316579483165792165792483792483165924831657831657924657924831",
		},
		{
			name: "seed 1",
			seed: 1,
			want: "691842357842357691357691842576918423918423576423576918235769184769184235184235769",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateWithoutShuffling(rand.New(rand.NewSource(tt.seed))).String(); got != tt.want {
				t.Errorf("generateWithoutShuffling()\ngot  = %v\nwant = %v", got, tt.want)
			}
		})
	}
}

func TestPuzzleTransformations(t *testing.T) {
	tests := []struct {
		name    string
		p       string
		fn      func(p app.PuzzleGenerator) error
		wantP   string
		wantErr bool
	}{
		// SWAP LINES

		{
			name: "swapLines a and b",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Horizontal, 0, 1)
			},
			wantP: "456789123123456789789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "swapLines b and a",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Horizontal, 1, 0)
			},
			wantP: "456789123123456789789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "swapLines a and c",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Horizontal, 0, 2)
			},
			wantP: "789123456456789123123456789891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "swapLines h and i",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Horizontal, 7, 8)
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345345678912912345678",
		},
		{
			name: "swapLines 1 and 2",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Vertical, 0, 1)
			},
			wantP: "213456789546789123879123456981234567324567891657891234768912345192345678435678912",
		},
		{
			name: "swapLines 2 and 1",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Vertical, 1, 0)
			},
			wantP: "213456789546789123879123456981234567324567891657891234768912345192345678435678912",
		},
		{
			name: "swapLines 8 and 9",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Vertical, 7, 8)
			},
			wantP: "123456798456789132789123465891234576234567819567891243678912354912345687345678921",
		},
		{
			name: "swapLines 200 and 1",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Vertical, 200, 1)
			},
			wantErr: true,
		},
		{
			name: "swapLines unknown direction",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.DirectionType(200), 0, 1)
			},
			wantErr: true,
		},
		{
			name: "swapLines 1 and 200",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Vertical, 1, 200)
			},
			wantErr: true,
		},
		{
			name: "swapLines 1 and 1",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapLines(app.Vertical, 1, 1)
			},
			wantErr: true,
		},

		// SWAP BIG LINES

		{
			name: "swapBigLines a-c and d-f",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Horizontal, 0, 1)
			},
			wantP: "891234567234567891567891234123456789456789123789123456678912345912345678345678912",
		},
		{
			name: "swapBigLines a-c and e-i",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Horizontal, 0, 2)
			},
			wantP: "678912345912345678345678912891234567234567891567891234123456789456789123789123456",
		},
		{
			name: "swapBigLines e-i and a-c",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Horizontal, 2, 0)
			},
			wantP: "678912345912345678345678912891234567234567891567891234123456789456789123789123456",
		},
		{
			name: "swapBigLines 1-3 and 4-6",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Vertical, 0, 1)
			},
			wantP: "456123789789456123123789456234891567567234891891567234912678345345912678678345912",
		},
		{
			name: "swapBigLines 7-9 and 4-6",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Vertical, 2, 1)
			},
			wantP: "123789456456123789789456123891567234234891567567234891678345912912678345345912678",
		},
		{
			name: "swapBigLines 10-12 and 4-6",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Vertical, 3, 1)
			},
			wantErr: true,
		},
		{
			name: "swapBigLines unknown direction",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.DirectionType(200), 0, 1)
			},
			wantErr: true,
		},
		{
			name: "swapBigLines 1-3 and 10-12",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Vertical, 0, 3)
			},
			wantErr: true,
		},
		{
			name: "swapBigLines a-c and a-c",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapBigLines(app.Horizontal, 0, 0)
			},
			wantErr: true,
		},

		// ROTATE

		{
			name: "rotate to 0",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Rotate(app.RotationType(0))
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "rotate to 90",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Rotate(app.RotateTo90)
			},
			wantP: "936714582825693471714582369693471258582369147471258936369147825258936714147825693",
		},
		{
			name: "rotate to 180",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Rotate(app.RotateTo180)
			},
			wantP: "219876543876543219543219876432198765198765432765432198654321987321987654987654321",
		},
		{
			name: "rotate to 270",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Rotate(app.RotateTo270)
			},
			wantP: "396528741417639852528741963639852174741963285852174396963285417174396528285417639",
		},
		{
			name: "rotate to 360",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return someErr(
					p.Rotate(app.RotateTo180),
					p.Rotate(app.RotateTo180),
				)
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "rotate to 360",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Rotate(app.RotationType(4))
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},

		// REFLECT

		{
			name: "reflect horizontal",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Reflect(app.ReflectHorizontal)
			},
			wantP: "987654321321987654654321987765432198198765432432198765543219876876543219219876543",
		},
		{
			name: "reflect horizontal double",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return someErr(
					p.Reflect(app.ReflectHorizontal),
					p.Reflect(app.ReflectHorizontal),
				)
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "reflect vertical",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Reflect(app.ReflectVertical)
			},
			wantP: "345678912912345678678912345567891234234567891891234567789123456456789123123456789",
		},
		{
			name: "reflect vertical double",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return someErr(
					p.Reflect(app.ReflectVertical),
					p.Reflect(app.ReflectVertical),
				)
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "reflect major diagonal",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Reflect(app.ReflectMajorDiagonal)
			},
			wantP: "147825693258936714369147825471258936582369147693471258714582369825693471936714582",
		},
		{
			name: "reflect major diagonal double",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return someErr(
					p.Reflect(app.ReflectMajorDiagonal),
					p.Reflect(app.ReflectMajorDiagonal),
				)
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "reflect minor diagonal",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Reflect(app.ReflectMinorDiagonal)
			},
			wantP: "285417639174396528963285417852174396741963285639852174528741963417639852396528741",
		},
		{
			name: "reflect minor diagonal double",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return someErr(
					p.Reflect(app.ReflectMinorDiagonal),
					p.Reflect(app.ReflectMinorDiagonal),
				)
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "reflect unknown reflection type",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.Reflect(app.ReflectionType(200))
			},
			wantErr: true,
		},

		// SWAP DIGITS

		{
			name: "swap digits 1 and 6",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapDigits(1, 6)
			},
			wantP: "623451789451789623789623451896234517234517896517896234178962345962345178345178962",
		},
		{
			name: "swap digits 6 and 1",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapDigits(6, 1)
			},
			wantP: "623451789451789623789623451896234517234517896517896234178962345962345178345178962",
		},
		{
			name: "swap digits 6 and 200 failed",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapDigits(6, 200)
			},
			wantErr: true,
		},
		{
			name: "swap digits 200 and 6 failed",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapDigits(200, 6)
			},
			wantErr: true,
		},
		{
			name: "swap digits 6 and 6 failed",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				return p.SwapDigits(6, 6)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParseGenerator(tt.p)
			if err != nil {
				t.Errorf("ParseGenerator() error: %v", err)
				return
			}
			err = tt.fn(p)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("fn() error = %v, wantErr is false", err)
					return
				}
				return
			} else {
				if tt.wantErr {
					t.Errorf("fn() error = <nil>, wantErr is true")
					return
				}
			}
			got := p.String()
			if got != tt.wantP {
				t.Errorf("Transformations()\ngot  = %s\nwant = %s", got, tt.wantP)
				return
			}
		})
	}
}

func TestForEach(t *testing.T) {
	const puzzle = `123456789456789123789123456891234567234567891567891234678912345912345678345678912`
	p, err := parse(puzzle)
	if err != nil {
		t.Fatal(err)
	}

	got, want := "", puzzle
	p.forEach(func(_ app.Point, val uint8, _ *bool) {
		got += strconv.Itoa(int(val))
	})
	if got != want {
		t.Errorf("forEach()\ngot  = %s\nwant = %s", got, want)
	}

	got, want = "", "123456789"
	p.forEachInRow(0, func(_ app.Point, val uint8, _ *bool) {
		got += strconv.Itoa(int(val))
	})
	if got != want {
		t.Errorf("forEachInRow()\ngot  = %s\nwant = %s", got, want)
	}

	got, want = "", "1475693"
	p.forEachInCol(0, func(_ app.Point, val uint8, _ *bool) {
		got += strconv.Itoa(int(val))
	}, 3, 4)
	if got != want {
		t.Errorf("forEachInCol()\ngot  = %s\nwant = %s", got, want)
	}

	got, want = "", "345678912"
	p.forEachInBox(app.Point{Row: 7, Col: 7}, func(_ app.Point, val uint8, _ *bool) {
		got += strconv.Itoa(int(val))
	})
	if got != want {
		t.Errorf("forEachInBox()\ngot  = %s\nwant = %s", got, want)
	}

	got, want = "", "(2; 2)"
	// skip the first two nines
	p.forEach(func(point app.Point, val uint8, stop *bool) {
		if val != 9 {
			return
		}
		got += fmt.Sprintf("(%d; %d)", point.Row, point.Col)
		*stop = true
	}, app.Point{Row: 0, Col: 8}, app.Point{Row: 1, Col: 5})
	if got != want {
		t.Errorf("forEach()\ngot  = %s\nwant = %s", got, want)
	}
}

func TestFindSimpleCandidates(t *testing.T) {
	const (
		puzzle = "400000938032094100095300240370609004529001673604703090957008300003900400240030709"
		want   = `{"a2":[1,6],"a3":[1,6],"a4":[1,2,5],"a5":[1,2,5,6,7],"a6":[2,5,6,7],"b1":[7,8],"b4":[5,8],"b8":[5,6],"b9":[5,6,7],"c1":[1,7,8],"c5":[1,6,7,8],"c6":[6,7],"c9":[6,7],"d3":[1,8],"d5":[2,5,8],"d7":[5,8],"d8":[1,2,5,8],"e4":[4,8],"e5":[4,8],"f2":[1,8],"f5":[2,5,8],"f7":[5,8],"f9":[1,2,5],"g4":[1,2,4],"g5":[1,2,4,6],"g8":[1,2,6],"g9":[1,2,6],"h1":[1,8],"h2":[1,6,8],"h5":[1,2,5,6,7],"h6":[2,5,6,7],"h8":[1,2,5,6,8],"h9":[1,2,5,6],"i3":[1,6,8],"i4":[1,5],"i6":[5,6],"i8":[1,5,6,8]}`
	)
	p, err := parse(puzzle)
	if err != nil {
		t.Fatal(err)
	}
	candidates := p.findSimpleCandidates()
	bts, err := json.Marshal(candidates)
	if err != nil {
		t.Errorf("puzzleCandidates.MarshalJSON() error = %v", err)
		return
	}
	if string(bts) != want {
		t.Errorf("puzzleCandidates.MarshalJSON()\ngot  = %s\nwant = %s", string(bts), want)
	}
}

func TestSolveSimpleSteps(t *testing.T) {
	tests := []struct {
		name  string
		p     string
		wantP string
	}{
		{
			// Naked Single
			name:  "Example Easiest Sudoku",
			p:     "...1.5...14....67..8...24...63.7..1.9.......3.1..9.52...72...8..26....35...4.9...",
			wantP: "672145398145983672389762451263574819958621743714398526597236184426817935831459267",
		},
		{
			// Hidden Single + Naked Single
			name:  "Example Gentle",
			p:     ".....4.284.6.....51...3.6.....3.1....87...14....7.9.....2.1...39.....5.767.4.....",
			wantP: "735164928426978315198532674249381756387256149561749832852617493914823567673495281",
		},
		/*{
			// Hidden Pair + Naked Triple + Hidden Single + Naked Single
			name:  "Example Moderate",
			p:     "72..96..3...2.5....8...4.2........6.1.65.38.7.4........3.8...9....7.2...2..43..18",
			wantP: "725196483463285971981374526372948165196523847548617239634851792819762354257439618",
		},*/
		/*{
			// Y-Wing + X-Wing + Naked Triple + Naked Pair + Hidden Single + Naked Single
			name:  "Example Tough",
			p:     "3.9...4..2..7.9....87......75..6.23.6..9.4..8.28.5..41......59....1.6..7..6...1.4",
			wantP: "369218475215749863487635912754861239631924758928357641173482596542196387896573124",
		},*/
		/*{
			// XY-Chain + X-Cycles + XYZ Wing + Simple Colouring + Y-Wing + X-Wing + Pointing Pair + Hidden Triple +
			//  Hidden Pair + Naked Triple + Naked Pair + Hidden Single + Naked Single
			name:  "Example Diabolical",
			p:     "...7.4..5.2..1..7.....8...2.9...625.6...7...8.532...1.4...9.....3..6..9.2..4.7...",
			wantP: "981724365324615879765983142197836254642571938853249716476398521538162497219457683",
		},*/
		{
			// Hidden Single + Naked Single
			name:  "Example Easy 17 Clue",
			p:     "....41....6....2...........32.6.........5..417...........2..3...48......5.1......",
			wantP: "872941563169573284453826197324617859986352741715498632697284315248135976531769428",
		},
		{
			// Naked Triple + Naked Pair + Hidden Single + Naked Single
			name:  "Example Naked Triples",
			p:     "...........19..5..56.31..9.1..6...28..4...7..27...4..3.4..68.35..2..59...........",
			wantP: "928547316431986572567312894195673428384251769276894153749168235612435987853729641",
		},

		// LESSON NAKED PAIRS, TRIPLES, QUADS

		{
			name:  "Strategy Lesson Naked Pair #1",
			p:     "4......38..2..41....53..24..7.6.9..4.2.....7.6..7.3.9..57..83....39..4..24......9",
			wantP: "461572938732894156895316247378629514529481673614753892957248361183967425246135789",
		},
		{
			name:  "Strategy Lesson Naked Pair #2",
			p:     ".8..9..3..3.........2.6.1.8.2.8..5..8..9.7..6..4..5.7.5.3.4.9.........1..1..5..2.",
			wantP: "486591732135278469972463158627814593851937246394625871563142987249786315718359624",
		},
		{
			name:  "Strategy Lesson Naked Triple #1",
			p:     ".7...8.29..2.....4854.2......83742.............32617......9.6122.....4..13.6...7.",
			wantP: "671438529392715864854926137518374296726859341943261785487593612269187453135642978",
		},
		{
			name:  "Strategy Lesson Naked Triple #2",
			p:     "2...1....6..8....93..6.7.54....56....4..8..6....47....73.1.4..59....5..1....2...7",
			wantP: "294513876675842319318697254129356748547289163863471592732164985986735421451928637",
		},
		/*{ // TODO naked quad
			name:  "Strategy Lesson Naked Quad",
			p:     "....3..86....2.........85..371....949.......54....76..2..7..8...3...5...7....4.3.",
			wantP: "142539786587621943693478521371856294968142375425397618214763859839215467756984132",
		},*/

		// LESSON HIDDEN PAIRS, TRIPLES, QUADS

		{
			name:  "Strategy Lesson Hidden Pair #1",
			p:     ".........9.46.7....768.41..3.97.1.8...8...3...5.3.87.2..75.261....4.32.8.........",
			wantP: "583219467914637825276854139349721586728965341651348792497582613165493278832176954",
		},
		/*{
			name:  "Strategy Lesson Hidden Pair #2", // TODO need Y-Wing
			p:     "72.4...3........47..1.768.2.1..39......8.1......26..8.2.968.4..34........6...3.75",
			wantP: "725498136986312547431576892812739654674851329593264781259687413347125968168943275",
		},*/
		{
			name:  "Strategy Lesson Hidden Triple",
			p:     ".........231.9.....65..31....8924...1...5...6...1367....93..57.....1.843.........",
			wantP: "894571632231698457765243198678924315143857926952136784489362571526719843317485269",
		},
		/*{ // TODO hidden quad
			name:  "Strategy Lesson Hidden Quad #1",
			p:     "65.....24...6.9....4.......57.4...61...5.1...31...2.85.......1....2.3...13.....98",
			wantP: "659387124721649853843125679572438961498561732316972485265894317987213546134756298",
		},
		{
			name:  "Strategy Lesson Hidden Quad #2",
			p:     "...5.....425.9...18...1..2.5.........19...46.........2.9..4...32...6.8.7.....16..",
			wantP: "971582346425693781863714529542136978319278465687459132196847253234965817758321694",
		},*/

		// POINTING PAIRS OR TRIPLES

		{
			name:  "Strategy Lesson Pointing Pair #1",
			p:     ".1.9.36......8....9.....5.7..2.1.43....4.2....64.7.2..7.1.....5....3......56.1.2.",
			wantP: "417953682256187943983246517872519436539462871164378259791824365628735194345691728",
		},
		{
			name:  "Strategy Lesson Pointing Pair #2",
			p:     ".32..61..41..........9.1...5...9...4.6.....7.3...2...5...5.8..........19..7...86.",
			wantP: "732456198419283756685971423528197634964835271371624985296518347843762519157349862",
		},
		{
			name:  "Strategy Lesson Pointing Triple",
			p:     "9...5....2..63...5..6..2.....31...7.....2.9...8...5......8..1..5...1...4....6...8",
			wantP: "931758246247631895856942317493186572165427983782395461624873159578219634319564728",
		},

		// PAIRS OR TRIPLES BOX/LINE REDUCTION

		/*{ // TODO Pair/Triple BLR
			name:  "Strategy Lesson Pair Box/Line Reduction",
			p:     ".16..78.3.9.8.....87...126..48...3..65...9.82.39...65..6.9...2..8...29369246..51.",
			wantP: "416527893592836147873491265148265379657319482239784651361958724785142936924673518",
		},
		{
			name:  "Strategy Lesson Triple Box/Line Reduction",
			p:     ".2.9437159.4...6..75.....4.5..48....2.....4534..352....42....81..5..426..9.2.85.4",
			wantP: "826943715934571628751826349563487192278619453419352876642735981385194267197268534",
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parse(tt.p)
			if err != nil {
				t.Fatal(err)
			}
			c := p.findSimpleCandidates()
			chanSteps := make(chan puzzleStep)
			go func() {
				changed, err := p.solve(&c, chanSteps)
				if err != nil {
					t.Error(err)
					return
				}
				if !changed {
					t.Errorf("solve() is not helped")
					return
				}
			}()
			for step := range chanSteps {
				t.Logf("step %s: %s", step.Strategy(), step.Description())
			}
			if got := p.String(); got != tt.wantP {
				t.Errorf("solve()\ngot  = %s\nwant = %s", got, tt.wantP)
				t.Logf("got:\n%s", c.debug(p))
				return
			}
		})
	}
}

func someErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
