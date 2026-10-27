package text

import "unicode/utf8"

// Distance returns the optimal string alignment distance between a and b:
// the fewest insertions, deletions, substitutions and swaps of two adjacent
// letters, where no letter is edited twice (restricted Damerau–Levenshtein,
// Damerau 1964, by the Wagner–Fischer dynamic program, 1974). It counts
// letters, not bytes.
func Distance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev2 := make([]int, len(rb)+1)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				cur[j] = min(cur[j], prev2[j-2]+1)
			}
		}
		prev2, prev, cur = prev, cur, prev2
	}
	return prev[len(rb)]
}

// Allowance returns how many edits a typed word of this length may be off
// by: none up to 3 letters, one up to 7, two beyond.
func Allowance(word string) int {
	switch n := utf8.RuneCountInString(word); {
	case n <= 3:
		return 0
	case n <= 7:
		return 1
	}
	return 2
}
