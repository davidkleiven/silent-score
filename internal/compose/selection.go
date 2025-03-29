package compose

import "strings"

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

func toSet(data []string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range data {
		result[item] = struct{}{}
	}
	return result
}
