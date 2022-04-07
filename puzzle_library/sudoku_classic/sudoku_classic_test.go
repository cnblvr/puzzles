package sudoku_classic

import (
	"github.com/cnblvr/puzzles/app"
	"math/rand"
	"testing"
)

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
		{
			name: "rotate to 0",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				p.Rotate(app.RotationType(0))
				return nil
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "rotate to 90",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				p.Rotate(app.RotateTo90)
				return nil
			},
			wantP: "936714582825693471714582369693471258582369147471258936369147825258936714147825693",
		},
		{
			name: "rotate to 180",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				p.Rotate(app.RotateTo180)
				return nil
			},
			wantP: "219876543876543219543219876432198765198765432765432198654321987321987654987654321",
		},
		{
			name: "rotate to 270",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				p.Rotate(app.RotateTo270)
				return nil
			},
			wantP: "396528741417639852528741963639852174741963285852174396963285417174396528285417639",
		},
		{
			name: "rotate to 360",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				p.Rotate(app.RotateTo180)
				p.Rotate(app.RotateTo180)
				return nil
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "rotate to 360",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleGenerator) error {
				p.Rotate(app.RotationType(4))
				return nil
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
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

func someErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
