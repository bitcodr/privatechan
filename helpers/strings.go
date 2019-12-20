//Package helpers ...
package helpers

import "sort"

func SortAndSearchInStrings(strings []string, find string) bool {
	sort.Strings(strings)
	index := sort.SearchStrings(strings, find)
	return strings[index] == find
}
