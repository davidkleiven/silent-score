package compose

import (
	"math"
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {
	data := "my-string,consists.of two, repeated,.delims"
	want := []string{"my-string", "consists", "of", "two", "repeated", "delims"}

	result := split(data)
	if slices.Compare(result, want) != 0 {
		t.Errorf("Wanted %v\ngot%v\n", want, result)
	}
}

func TestNgram(t *testing.T) {
	data := "abcdefg"
	trigrams := []string{"abc", "bcd", "cde", "def", "efg"}
	result := ngram(data, 3)

	if len(trigrams) != len(result) {
		t.Errorf("Wanted %d trigrams got %d", len(trigrams), len(result))
	}

	for _, trigram := range trigrams {
		if _, ok := result[trigram]; !ok {
			t.Errorf("%s not part of %v", trigram, result)
			break
		}
	}
}

func TestJaccard(t *testing.T) {
	tol := 1e-6
	for i, test := range []struct {
		data1 []string
		data2 []string
		want  float64
	}{
		{
			data1: []string{"a", "b"},
			data2: []string{"b", "a"},
			want:  1.0,
		},
		{
			data1: []string{},
			data2: []string{},
			want:  1.0,
		},
		{
			data1: []string{"a"},
			data2: []string{"a", "b"},
			want:  0.5,
		},
		{
			data1: []string{"c"},
			data2: []string{"a", "b"},
			want:  0.0,
		},
	} {
		jaccardIndex := jaccard(toSet(test.data1), toSet(test.data2))
		if math.Abs(jaccardIndex-test.want) > tol {
			t.Errorf("Test #%d: Wanted %f got %f", i, test.want, jaccardIndex)
		}
	}
}

func TestNormalize(t *testing.T) {
	data := "aAb"
	normalized := normalize(data)
	want := "aab"
	if normalized != want {
		t.Errorf("Wanted %s got %s", want, normalized)
	}
}

func TestNumOverlapping(t *testing.T) {
	target := map[string]struct{}{
		"a": {},
		"b": {},
	}
	candidate := map[string]struct{}{
		"b": {},
		"d": {},
		"e": {},
	}

	if r := numCommonElements(target, candidate); r != 1 {
		t.Errorf("Wanted 1 overlapping god %d", r)
	}

}

func TestOrderPiece(t *testing.T) {
	desc := "agitato, shark in the water"
	pieces := []string{
		"fish water andante",
		"water agitato",
		"written by Cole Porter",
	}

	result := orderPieces(desc, pieces)
	wantIndexes := []int{1, 0, 2}
	gotIndex := make([]int, 3)
	for i, v := range result {
		gotIndex[i] = v.Index
	}

	if slices.Compare(gotIndex, wantIndexes) != 0 {
		t.Errorf("Wanted %v got %v", wantIndexes, gotIndex)
	}
}
