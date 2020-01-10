//Package helpers ...
package helpers

import (
	"log"
	"regexp"
	"sort"
)

func SortAndSearchInStrings(strings []string, find string) bool {
	sort.Strings(strings)
	index := sort.SearchStrings(strings, find)
	return strings[index] == find
}

func ClearString(text string) string {
	data, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Println(err)
		return "-"
	}
	return data.ReplaceAllString(text, " ")
}
