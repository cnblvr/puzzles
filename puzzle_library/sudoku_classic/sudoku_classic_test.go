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
			p, err := Parse(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr is false", err)
					return
				}
				return
			} else {
				if tt.wantErr {
					t.Errorf("Parse() error = <nil>, wantErr is true")
					return
				}
			}
			out := p.String()
			if out != tt.wantOut {
				t.Errorf("Parse().String() got = %s, want %s", out, tt.wantOut)
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
				t.Errorf("generateWithoutShuffling() got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_puzzle_rotate(t *testing.T) {
	tests := []struct {
		name  string
		p     string
		fn    func(p app.PuzzleAssistant)
		wantP string
	}{
		{
			name: "rotate to 0",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleAssistant) {
				p.Rotate(app.RotationType(0))
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "rotate to 90",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleAssistant) {
				p.Rotate(app.RotateTo90)
			},
			wantP: "936714582825693471714582369693471258582369147471258936369147825258936714147825693",
		},
		{
			name: "rotate to 180",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleAssistant) {
				p.Rotate(app.RotateTo180)
			},
			wantP: "219876543876543219543219876432198765198765432765432198654321987321987654987654321",
		},
		{
			name: "rotate to 270",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleAssistant) {
				p.Rotate(app.RotateTo270)
			},
			wantP: "396528741417639852528741963639852174741963285852174396963285417174396528285417639",
		},
		{
			name: "rotate to 360",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleAssistant) {
				p.Rotate(app.RotateTo180)
				p.Rotate(app.RotateTo180)
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
		{
			name: "rotate to 360",
			p:    "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
			fn: func(p app.PuzzleAssistant) {
				p.Rotate(app.RotationType(4))
			},
			wantP: "123456789456789123789123456891234567234567891567891234678912345912345678345678912",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Parse(tt.p)
			if err != nil {
				t.Errorf("p invalid: %v", err)
				return
			}
			tt.fn(p)
			got := p.String()
			if got != tt.wantP {
				t.Errorf("rotate() got = %s, want %s", got, tt.wantP)
				return
			}
		})
	}
}
