package compose

import (
	"cmp"
	"slices"
	"strings"
)

func normalize(data string) string {
	return strings.ToLower(data)
}

func split(data string) []string {
	return strings.FieldsFunc(data, func(c rune) bool {
		return c == ',' || c == ' ' || c == '.'
	})
}

func ngram(data string, num int) map[string]struct{} {
	result := make(map[string]struct{})
	for i := range len(data) - num + 1 {
		result[data[i:i+num]] = struct{}{}
	}
	return result
}

func jaccard(data1 map[string]struct{}, data2 map[string]struct{}) float64 {
	if len(data1) == 0 && len(data2) == 0 {
		return 1.0
	}

	intersection := make(map[string]struct{})
	union := make(map[string]struct{})
	for item := range data1 {
		if _, ok := data2[item]; ok {
			intersection[item] = struct{}{}
		}
		union[item] = struct{}{}
	}

	for item := range data2 {
		union[item] = struct{}{}
	}

	return float64(len(intersection)) / float64(len(union))
}

func numCommonElements(target map[string]struct{}, candidate map[string]struct{}) int {
	numCommon := 0
	for v := range candidate {
		if _, ok := target[v]; ok {
			numCommon += 1
		}
	}
	return numCommon
}

func toSet(data []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range data {
		result[item] = struct{}{}
	}
	return result
}

type Score struct {
	Similarity int
	Index      int
}

func orderPieces(desc string, pieceText []string) []Score {
	targetTokens := ngram(normalize(desc), 3)

	similarity := make([]Score, len(pieceText))
	for i, text := range pieceText {
		similarity[i] = Score{
			Similarity: numCommonElements(ngram(normalize(text), 3), targetTokens),
			Index:      i,
		}
	}
	slices.SortStableFunc(similarity, func(s1, s2 Score) int {
		return -cmp.Compare(s1.Similarity, s2.Similarity)
	})
	return similarity
}
