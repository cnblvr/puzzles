package sudoku_classic

import (
	"bytes"
	"github.com/cnblvr/puzzles/app"
	"reflect"
	"testing"
)

func TestPuzzleCandidates_encode(t *testing.T) {
	tests := []struct {
		name string
		in   puzzleCandidates
		want string
	}{
		{
			name: "empty",
			in:   newPuzzleCandidates(false),
			want: `{}`,
		},
		{
			name: "filled",
			in:   newPuzzleCandidates(true),
			want: `{"base":{"a1":[1,2,3,4,5,6,7,8,9],"a2":[1,2,3,4,5,6,7,8,9],"a3":[1,2,3,4,5,6,7,8,9],"a4":[1,2,3,4,5,6,7,8,9],"a5":[1,2,3,4,5,6,7,8,9],"a6":[1,2,3,4,5,6,7,8,9],"a7":[1,2,3,4,5,6,7,8,9],"a8":[1,2,3,4,5,6,7,8,9],"a9":[1,2,3,4,5,6,7,8,9],"b1":[1,2,3,4,5,6,7,8,9],"b2":[1,2,3,4,5,6,7,8,9],"b3":[1,2,3,4,5,6,7,8,9],"b4":[1,2,3,4,5,6,7,8,9],"b5":[1,2,3,4,5,6,7,8,9],"b6":[1,2,3,4,5,6,7,8,9],"b7":[1,2,3,4,5,6,7,8,9],"b8":[1,2,3,4,5,6,7,8,9],"b9":[1,2,3,4,5,6,7,8,9],"c1":[1,2,3,4,5,6,7,8,9],"c2":[1,2,3,4,5,6,7,8,9],"c3":[1,2,3,4,5,6,7,8,9],"c4":[1,2,3,4,5,6,7,8,9],"c5":[1,2,3,4,5,6,7,8,9],"c6":[1,2,3,4,5,6,7,8,9],"c7":[1,2,3,4,5,6,7,8,9],"c8":[1,2,3,4,5,6,7,8,9],"c9":[1,2,3,4,5,6,7,8,9],"d1":[1,2,3,4,5,6,7,8,9],"d2":[1,2,3,4,5,6,7,8,9],"d3":[1,2,3,4,5,6,7,8,9],"d4":[1,2,3,4,5,6,7,8,9],"d5":[1,2,3,4,5,6,7,8,9],"d6":[1,2,3,4,5,6,7,8,9],"d7":[1,2,3,4,5,6,7,8,9],"d8":[1,2,3,4,5,6,7,8,9],"d9":[1,2,3,4,5,6,7,8,9],"e1":[1,2,3,4,5,6,7,8,9],"e2":[1,2,3,4,5,6,7,8,9],"e3":[1,2,3,4,5,6,7,8,9],"e4":[1,2,3,4,5,6,7,8,9],"e5":[1,2,3,4,5,6,7,8,9],"e6":[1,2,3,4,5,6,7,8,9],"e7":[1,2,3,4,5,6,7,8,9],"e8":[1,2,3,4,5,6,7,8,9],"e9":[1,2,3,4,5,6,7,8,9],"f1":[1,2,3,4,5,6,7,8,9],"f2":[1,2,3,4,5,6,7,8,9],"f3":[1,2,3,4,5,6,7,8,9],"f4":[1,2,3,4,5,6,7,8,9],"f5":[1,2,3,4,5,6,7,8,9],"f6":[1,2,3,4,5,6,7,8,9],"f7":[1,2,3,4,5,6,7,8,9],"f8":[1,2,3,4,5,6,7,8,9],"f9":[1,2,3,4,5,6,7,8,9],"g1":[1,2,3,4,5,6,7,8,9],"g2":[1,2,3,4,5,6,7,8,9],"g3":[1,2,3,4,5,6,7,8,9],"g4":[1,2,3,4,5,6,7,8,9],"g5":[1,2,3,4,5,6,7,8,9],"g6":[1,2,3,4,5,6,7,8,9],"g7":[1,2,3,4,5,6,7,8,9],"g8":[1,2,3,4,5,6,7,8,9],"g9":[1,2,3,4,5,6,7,8,9],"h1":[1,2,3,4,5,6,7,8,9],"h2":[1,2,3,4,5,6,7,8,9],"h3":[1,2,3,4,5,6,7,8,9],"h4":[1,2,3,4,5,6,7,8,9],"h5":[1,2,3,4,5,6,7,8,9],"h6":[1,2,3,4,5,6,7,8,9],"h7":[1,2,3,4,5,6,7,8,9],"h8":[1,2,3,4,5,6,7,8,9],"h9":[1,2,3,4,5,6,7,8,9],"i1":[1,2,3,4,5,6,7,8,9],"i2":[1,2,3,4,5,6,7,8,9],"i3":[1,2,3,4,5,6,7,8,9],"i4":[1,2,3,4,5,6,7,8,9],"i5":[1,2,3,4,5,6,7,8,9],"i6":[1,2,3,4,5,6,7,8,9],"i7":[1,2,3,4,5,6,7,8,9],"i8":[1,2,3,4,5,6,7,8,9],"i9":[1,2,3,4,5,6,7,8,9]}}`,
		},
		{
			name: "simple",
			in: func() (p puzzleCandidates) {
				p[0][0] = newCellCandidatesWith(1, 3, 5, 7, 9)
				p[8][8] = newCellCandidatesWith(2, 4, 6, 8)
				return
			}(),
			want: `{"base":{"a1":[1,3,5,7,9],"i9":[2,4,6,8]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.encode()
			if got != tt.want {
				t.Errorf("encode() got = %v, want = %v", got, tt.want)
				return
			}
		})
	}
}

func TestPuzzleCandidates_encodeOnlyChanges(t *testing.T) {
	type args struct {
		base    puzzleCandidates
		changes puzzleCandidates
	}
	tests := []struct {
		name string
		in   args
		want string
	}{
		{
			name: "empty",
			in: args{
				base:    newPuzzleCandidates(false),
				changes: newPuzzleCandidates(false),
			},
			want: `{}`,
		},
		{
			name: "simple",
			in: func() args {
				base := newPuzzleCandidates(false)
				base[0][0] = newCellCandidatesWith(1, 3, 5, 7, 9)
				base[8][8] = newCellCandidatesWith(2, 4, 6, 8)

				changes := base.clone()
				changes[0][0].delete(3, 9)
				changes[8][8].delete(2, 4)
				changes[8][8].add(2)
				changes[8][8].add(3)
				changes[5][5].add(1, 2, 3)

				return args{
					base:    base,
					changes: changes,
				}
			}(),
			want: `{"add":{"f6":[1,2,3],"i9":[3]},"del":{"a1":[3,9],"i9":[4]}}`,
		},
		{
			name: "only add",
			in: func() args {
				base := newPuzzleCandidates(false)
				base[0][0] = newCellCandidatesWith(1, 3, 5, 7, 9)
				base[8][8] = newCellCandidatesWith(2, 4, 6, 8)

				changes := base.clone()
				changes[8][8].add(2)
				changes[8][8].add(3)
				changes[5][5].add(1, 2, 3)

				return args{
					base:    base,
					changes: changes,
				}
			}(),
			want: `{"add":{"f6":[1,2,3],"i9":[3]}}`,
		},
		{
			name: "only delete",
			in: func() args {
				base := newPuzzleCandidates(false)
				base[0][0] = newCellCandidatesWith(1, 3, 5, 7, 9)
				base[8][8] = newCellCandidatesWith(2, 4, 6, 8)

				changes := base.clone()
				changes[0][0].delete(3, 9)
				changes[8][8].delete(2, 4)
				changes[5][5].delete(1, 2, 3)

				return args{
					base:    base,
					changes: changes,
				}
			}(),
			want: `{"del":{"a1":[3,9],"i9":[2,4]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.changes.encodeOnlyChanges(tt.in.base)
			if got != tt.want {
				t.Errorf("encodeOnlyChanges() got = %v, want = %v", got, tt.want)
				return
			}
		})
	}
}

func Test_decodeCandidates(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    puzzleCandidates
		wantErr bool
	}{
		{
			name: "empty",
			in:   `{}`,
			want: newPuzzleCandidates(false),
		},
		{
			name: "filled",
			in:   `{"base":{"a1":[1,2,3,4,5,6,7,8,9],"a2":[1,2,3,4,5,6,7,8,9],"a3":[1,2,3,4,5,6,7,8,9],"a4":[1,2,3,4,5,6,7,8,9],"a5":[1,2,3,4,5,6,7,8,9],"a6":[1,2,3,4,5,6,7,8,9],"a7":[1,2,3,4,5,6,7,8,9],"a8":[1,2,3,4,5,6,7,8,9],"a9":[1,2,3,4,5,6,7,8,9],"b1":[1,2,3,4,5,6,7,8,9],"b2":[1,2,3,4,5,6,7,8,9],"b3":[1,2,3,4,5,6,7,8,9],"b4":[1,2,3,4,5,6,7,8,9],"b5":[1,2,3,4,5,6,7,8,9],"b6":[1,2,3,4,5,6,7,8,9],"b7":[1,2,3,4,5,6,7,8,9],"b8":[1,2,3,4,5,6,7,8,9],"b9":[1,2,3,4,5,6,7,8,9],"c1":[1,2,3,4,5,6,7,8,9],"c2":[1,2,3,4,5,6,7,8,9],"c3":[1,2,3,4,5,6,7,8,9],"c4":[1,2,3,4,5,6,7,8,9],"c5":[1,2,3,4,5,6,7,8,9],"c6":[1,2,3,4,5,6,7,8,9],"c7":[1,2,3,4,5,6,7,8,9],"c8":[1,2,3,4,5,6,7,8,9],"c9":[1,2,3,4,5,6,7,8,9],"d1":[1,2,3,4,5,6,7,8,9],"d2":[1,2,3,4,5,6,7,8,9],"d3":[1,2,3,4,5,6,7,8,9],"d4":[1,2,3,4,5,6,7,8,9],"d5":[1,2,3,4,5,6,7,8,9],"d6":[1,2,3,4,5,6,7,8,9],"d7":[1,2,3,4,5,6,7,8,9],"d8":[1,2,3,4,5,6,7,8,9],"d9":[1,2,3,4,5,6,7,8,9],"e1":[1,2,3,4,5,6,7,8,9],"e2":[1,2,3,4,5,6,7,8,9],"e3":[1,2,3,4,5,6,7,8,9],"e4":[1,2,3,4,5,6,7,8,9],"e5":[1,2,3,4,5,6,7,8,9],"e6":[1,2,3,4,5,6,7,8,9],"e7":[1,2,3,4,5,6,7,8,9],"e8":[1,2,3,4,5,6,7,8,9],"e9":[1,2,3,4,5,6,7,8,9],"f1":[1,2,3,4,5,6,7,8,9],"f2":[1,2,3,4,5,6,7,8,9],"f3":[1,2,3,4,5,6,7,8,9],"f4":[1,2,3,4,5,6,7,8,9],"f5":[1,2,3,4,5,6,7,8,9],"f6":[1,2,3,4,5,6,7,8,9],"f7":[1,2,3,4,5,6,7,8,9],"f8":[1,2,3,4,5,6,7,8,9],"f9":[1,2,3,4,5,6,7,8,9],"g1":[1,2,3,4,5,6,7,8,9],"g2":[1,2,3,4,5,6,7,8,9],"g3":[1,2,3,4,5,6,7,8,9],"g4":[1,2,3,4,5,6,7,8,9],"g5":[1,2,3,4,5,6,7,8,9],"g6":[1,2,3,4,5,6,7,8,9],"g7":[1,2,3,4,5,6,7,8,9],"g8":[1,2,3,4,5,6,7,8,9],"g9":[1,2,3,4,5,6,7,8,9],"h1":[1,2,3,4,5,6,7,8,9],"h2":[1,2,3,4,5,6,7,8,9],"h3":[1,2,3,4,5,6,7,8,9],"h4":[1,2,3,4,5,6,7,8,9],"h5":[1,2,3,4,5,6,7,8,9],"h6":[1,2,3,4,5,6,7,8,9],"h7":[1,2,3,4,5,6,7,8,9],"h8":[1,2,3,4,5,6,7,8,9],"h9":[1,2,3,4,5,6,7,8,9],"i1":[1,2,3,4,5,6,7,8,9],"i2":[1,2,3,4,5,6,7,8,9],"i3":[1,2,3,4,5,6,7,8,9],"i4":[1,2,3,4,5,6,7,8,9],"i5":[1,2,3,4,5,6,7,8,9],"i6":[1,2,3,4,5,6,7,8,9],"i7":[1,2,3,4,5,6,7,8,9],"i8":[1,2,3,4,5,6,7,8,9],"i9":[1,2,3,4,5,6,7,8,9]}}`,
			want: newPuzzleCandidates(true),
		},
		{
			name: "simple",
			in:   `{"base":{"a1":[1,3,5,7,9],"i9":[2,4,6,8]}}`,
			want: func() (p puzzleCandidates) {
				p = newPuzzleCandidates(false)
				p[0][0] = newCellCandidatesWith(1, 3, 5, 7, 9)
				p[8][8] = newCellCandidatesWith(2, 4, 6, 8)
				return
			}(),
		},
		{
			name:    "error parse json",
			in:      `{"base":{"a1":[1,3,5,7,9],"i9":[2,4,6,8`,
			wantErr: true,
		},
		{
			name:    "error parse point",
			in:      `{"base":{"w5":[1,2,3]}}`,
			wantErr: true,
		},
		{
			name:    "invalid candidates",
			in:      `{"base":{"a1":[1,2,12345]}}`,
			wantErr: true,
		},
		{
			name:    "invalid candidates",
			in:      `{"base":{"a1":[1,2,-123]}}`,
			wantErr: true,
		},
		{
			name:    "invalid candidates",
			in:      `{"base":{"a1":[1,2,123]}}`,
			wantErr: true,
		},
		{
			name:    "invalid candidates",
			in:      `{"base":{"a1":[1,2,10]}}`,
			wantErr: true,
		},
		{
			name:    "invalid candidates",
			in:      `{"base":{"a1":[1,2,0]}}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeCandidates(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("decodeCandidates() got err = %v, want err %v", err, tt.wantErr)
				}
				return
			} else {
				if tt.wantErr {
					t.Errorf("decodeCandidates() got err = %v, want err %v", err, tt.wantErr)
					return
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeCandidates() got = %v, want = %v", got, tt.want)
				return
			}
		})
	}
}

func TestCellCandidatesIntersection(t *testing.T) {
	tests := []struct {
		name string
		a    cellCandidates
		b    cellCandidates
		want []uint8
	}{
		{
			name: "{}⋂{}",
			a:    newCellCandidatesWith(),
			b:    newCellCandidatesWith(),
			want: []uint8{},
		},
		{
			name: "{1}⋂{}",
			a:    newCellCandidatesWith(1),
			b:    newCellCandidatesWith(),
			want: []uint8{},
		},
		{
			name: "{1}⋂{1}",
			a:    newCellCandidatesWith(1),
			b:    newCellCandidatesWith(1),
			want: []uint8{1},
		},
		{
			name: "{1,2,3}⋂{1,2,3}",
			a:    newCellCandidatesWith(1, 2, 3),
			b:    newCellCandidatesWith(1, 2, 3),
			want: []uint8{1, 2, 3},
		},
		{
			name: "{1,2,3}⋂{4,5,6}",
			a:    newCellCandidatesWith(1, 2, 3),
			b:    newCellCandidatesWith(4, 5, 6),
			want: []uint8{},
		},
		{
			name: "{1,2,3}⋂{2,3,4}",
			a:    newCellCandidatesWith(1, 2, 3),
			b:    newCellCandidatesWith(2, 3, 4),
			want: []uint8{2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.intersection(tt.b).slice()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("intersection() got = %v, want = %v", got, tt.want)
				return
			}
		})
	}
}

func TestCellCandidatesComplement(t *testing.T) {
	tests := []struct {
		name string
		a    cellCandidates
		b    cellCandidates
		want []uint8
	}{
		{
			name: "{}\\{}",
			a:    newCellCandidatesWith(),
			b:    newCellCandidatesWith(),
			want: []uint8{},
		},
		{
			name: "{1}\\{}",
			a:    newCellCandidatesWith(1),
			b:    newCellCandidatesWith(),
			want: []uint8{1},
		},
		{
			name: "{1}\\{1}",
			a:    newCellCandidatesWith(1),
			b:    newCellCandidatesWith(1),
			want: []uint8{},
		},
		{
			name: "{}\\{1,2,3}",
			a:    newCellCandidatesWith(),
			b:    newCellCandidatesWith(1, 2, 3),
			want: []uint8{},
		},
		{
			name: "{1}\\{1,2,3}",
			a:    newCellCandidatesWith(1),
			b:    newCellCandidatesWith(1, 2, 3),
			want: []uint8{},
		},
		{
			name: "{1,2,3}\\{1,2,3}",
			a:    newCellCandidatesWith(1, 2, 3),
			b:    newCellCandidatesWith(1, 2, 3),
			want: []uint8{},
		},
		{
			name: "{1,2,3}\\{1}",
			a:    newCellCandidatesWith(1, 2, 3),
			b:    newCellCandidatesWith(1),
			want: []uint8{2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.complement(tt.b).slice()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("complement() got = %v, want = %v", got, tt.want)
				return
			}
		})
	}
}

func TestCellCandidatesDeleteExcept(t *testing.T) {
	tests := []struct {
		name    string
		a       cellCandidates
		b       []uint8
		wantRet bool
		want    []uint8
	}{
		{
			name:    "{} delete expect {}",
			a:       newCellCandidatesWith(),
			b:       []uint8{},
			wantRet: false,
			want:    []uint8{},
		},
		{
			name:    "{1} delete expect {}",
			a:       newCellCandidatesWith(1),
			b:       []uint8{},
			wantRet: true,
			want:    []uint8{},
		},
		{
			name:    "{} delete expect {1}",
			a:       newCellCandidatesWith(),
			b:       []uint8{1},
			wantRet: false,
			want:    []uint8{},
		},
		{
			name:    "{1,2,3,4,5,6,7,8,9} delete expect {1,2,3}",
			a:       newCellCandidatesWith(1, 2, 3, 4, 5, 6, 7, 8, 9),
			b:       []uint8{1, 2, 3},
			wantRet: true,
			want:    []uint8{1, 2, 3},
		},
		{
			name:    "{1,2,3} delete expect {1,2,3}",
			a:       newCellCandidatesWith(1, 2, 3),
			b:       []uint8{1, 2, 3},
			wantRet: false,
			want:    []uint8{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet := tt.a.deleteExcept(tt.b...)
			got := tt.a.slice()
			if gotRet != tt.wantRet {
				t.Errorf("deleteExcept() got return = %v, want return = %v", gotRet, tt.wantRet)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("deleteExcept() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_TODO_Delete(t *testing.T) {
	a := newCellCandidatesWith(2, 3, 4, 8)
	b := newCellCandidatesWith(2, 3, 5)
	other := []cellCandidates{
		newCellCandidatesWith(1, 5),
		newCellCandidatesWith(4, 8),
		newCellCandidatesWith(),
		newCellCandidatesWith(1, 4, 8, 9),
		newCellCandidatesWith(1, 9),
		newCellCandidatesWith(),
		newCellCandidatesWith(8, 5),
	}
	complement := a.intersection(b)
	for _, o := range other {
		complement = complement.complement(o)
	}
	t.Logf("%v", complement.slice())
}

func TestBoxIdFrom(t *testing.T) {
	tests := []struct {
		name  string
		point app.Point
		want  uint8
	}{
		{
			name:  "a1",
			point: app.Point{Row: 0, Col: 0},
			want:  1,
		},
		{
			name:  "b1",
			point: app.Point{Row: 1, Col: 0},
			want:  1,
		},
		{
			name:  "b9",
			point: app.Point{Row: 1, Col: 8},
			want:  3,
		},
		{
			name:  "c4",
			point: app.Point{Row: 2, Col: 3},
			want:  2,
		},
		{
			name:  "d3",
			point: app.Point{Row: 3, Col: 2},
			want:  4,
		},
		{
			name:  "d5",
			point: app.Point{Row: 3, Col: 4},
			want:  5,
		},
		{
			name:  "e5",
			point: app.Point{Row: 4, Col: 4},
			want:  5,
		},
		{
			name:  "f8",
			point: app.Point{Row: 5, Col: 7},
			want:  6,
		},
		{
			name:  "h1",
			point: app.Point{Row: 7, Col: 0},
			want:  7,
		},
		{
			name:  "i9",
			point: app.Point{Row: 8, Col: 8},
			want:  9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BoxIdFrom(tt.point)
			if got != tt.want {
				t.Errorf("BoxIdFrom() got = %d, want = %d", got, tt.want)
				return
			}
		})
	}
}
