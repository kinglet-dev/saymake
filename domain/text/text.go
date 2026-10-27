// Package text turns a description into tokens and measures how far apart
// two words are, for fuzzy matching.
package text

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Limits on a description (see the threat model in DESIGN.md).
const (
	MaxChars  = 500
	MaxTokens = 60
	maxWhole  = 4 // digits before the point
	maxFrac   = 2 // digits after the point
)

// Token is a word or a number from a description.
type Token struct {
	Text       string
	Number     float64
	IsNumber   bool
	AfterBreak bool // punctuation (, ; : . ! ?) comes between it and the token before
}

// Tokenize splits a description into lowercase words and numbers. Letters
// (any script) make words; apostrophes inside a word are dropped; digits with
// at most one decimal point make numbers; a double quote straight after a
// number means inches; everything else separates tokens.
func Tokenize(s string) ([]Token, error) {
	if !utf8.ValidString(s) {
		return nil, errors.New("the description is not valid text")
	}
	rs := []rune(s)
	if len(rs) > MaxChars {
		return nil, fmt.Errorf("a description can be at most %d characters", MaxChars)
	}
	var out []Token
	brk := false
	for i := 0; i < len(rs); {
		var t Token
		var err error
		switch r := rs[i]; {
		case isLetter(r):
			t, i = word(rs, i)
		case dimensions(rs, i):
			t, i = Token{Text: string(rs[i]) + "d"}, i+2
		case unicode.IsDigit(r):
			t, i, err = number(rs, i)
			if err != nil {
				return nil, err
			}
		case isQuote(r) && i > 0 && unicode.IsDigit(rs[i-1]):
			t, i = Token{Text: "in"}, i+1
		default:
			brk = brk || strings.ContainsRune(",;:.!?", rs[i])
			i++
			continue
		}
		t.AfterBreak, brk = brk && len(out) > 0, false
		out = append(out, t)
		if len(out) > MaxTokens {
			return nil, fmt.Errorf("a description can be at most %d words", MaxTokens)
		}
	}
	return out, nil
}

func word(rs []rune, i int) (Token, int) {
	var w []rune
	for ; i < len(rs); i++ {
		r := rs[i]
		if isApostrophe(r) && len(w) > 0 && i+1 < len(rs) && isLetter(rs[i+1]) {
			continue
		}
		if !isLetter(r) {
			break
		}
		w = append(w, unicode.ToLower(r))
	}
	return Token{Text: string(w)}, i
}

func number(rs []rune, i int) (Token, int, error) {
	start, whole, frac, point := i, 0, 0, false
	for ; i < len(rs); i++ {
		r := rs[i]
		switch {
		case unicode.IsDigit(r) && point:
			frac++
		case unicode.IsDigit(r):
			whole++
		case r == '.' && !point && i+1 < len(rs) && unicode.IsDigit(rs[i+1]):
			point = true
		default:
			return finish(rs[start:i], whole, frac, i)
		}
	}
	return finish(rs[start:i], whole, frac, i)
}

func finish(digits []rune, whole, frac, next int) (Token, int, error) {
	lit := string(digits)
	if whole > maxWhole || frac > maxFrac {
		return Token{}, 0, fmt.Errorf("%q has too many digits; use at most %d before the point and %d after", lit, maxWhole, maxFrac)
	}
	// lit is digits with at most one point between digits, so it parses.
	v, _ := strconv.ParseFloat(lit, 64)
	return Token{Text: lit, Number: v, IsNumber: true}, next, nil
}

// dimensions reports "2d" or "3d" (any case) as a word on its own at i.
func dimensions(rs []rune, i int) bool {
	return (rs[i] == '2' || rs[i] == '3') && i+1 < len(rs) && unicode.ToLower(rs[i+1]) == 'd' &&
		(i+2 == len(rs) || !isLetter(rs[i+2]))
}

func isLetter(r rune) bool     { return unicode.IsLetter(r) || unicode.Is(unicode.Mn, r) }
func isApostrophe(r rune) bool { return r == '\'' || r == '’' }
func isQuote(r rune) bool      { return r == '"' || r == '”' || r == '″' }
