package sudoku_classic

import (
	"bytes"
	"testing"
)

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
