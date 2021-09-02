package queryparser

import (
	"sort"
	"strings"
)

func Matches(data, query string) bool {
	dataWords := strings.Split(data, " ")
	queryWords := strings.Split(query, " ")
	sort.Strings(dataWords)
	for _, qw := range queryWords {
		if qw == "" {
			continue
		}
		if contains(dataWords, qw) {
			return true
		}
	}
	return false
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}
