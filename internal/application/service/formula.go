package service

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// evalFormula evaluates a numeric expression using + - * / ( ), numeric literals, and dotted paths
// (e.g. metadata.counts_a, event.data.amount). Paths are resolved with getByPath on the same
// context as rule conditions. Division by zero yields 0.
func evalFormula(expr string, ctx map[string]interface{}) (float64, error) {
	expr = strings.TrimSpace(expr)
	l := &formulaLexer{s: expr, i: 0}
	l.skipSpace()
	if l.i >= len(l.s) {
		return 0, fmt.Errorf("empty formula")
	}
	v, err := parseFormulaExpr(l, ctx)
	if err != nil {
		return 0, err
	}
	l.skipSpace()
	if l.i < len(l.s) {
		return 0, fmt.Errorf("unexpected trailing input")
	}
	return v, nil
}

type formulaLexer struct {
	s string
	i int
}

func (l *formulaLexer) skipSpace() {
	for l.i < len(l.s) && unicode.IsSpace(rune(l.s[l.i])) {
		l.i++
	}
}

func parseFormulaExpr(l *formulaLexer, ctx map[string]interface{}) (float64, error) {
	left, err := parseFormulaTerm(l, ctx)
	if err != nil {
		return 0, err
	}
	for {
		l.skipSpace()
		if l.i >= len(l.s) {
			return left, nil
		}
		op := l.s[l.i]
		if op != '+' && op != '-' {
			return left, nil
		}
		l.i++
		right, err := parseFormulaTerm(l, ctx)
		if err != nil {
			return 0, err
		}
		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
}

func parseFormulaTerm(l *formulaLexer, ctx map[string]interface{}) (float64, error) {
	left, err := parseFormulaFactor(l, ctx)
	if err != nil {
		return 0, err
	}
	for {
		l.skipSpace()
		if l.i >= len(l.s) {
			return left, nil
		}
		op := l.s[l.i]
		if op != '*' && op != '/' {
			return left, nil
		}
		l.i++
		right, err := parseFormulaFactor(l, ctx)
		if err != nil {
			return 0, err
		}
		if op == '*' {
			left *= right
		} else if right == 0 {
			left = 0
		} else {
			left /= right
		}
	}
}

func parseFormulaFactor(l *formulaLexer, ctx map[string]interface{}) (float64, error) {
	l.skipSpace()
	if l.i >= len(l.s) {
		return 0, fmt.Errorf("unexpected end of formula")
	}
	if l.s[l.i] == '-' {
		l.i++
		v, err := parseFormulaFactor(l, ctx)
		return -v, err
	}
	if l.s[l.i] == '(' {
		l.i++
		v, err := parseFormulaExpr(l, ctx)
		if err != nil {
			return 0, err
		}
		l.skipSpace()
		if l.i >= len(l.s) || l.s[l.i] != ')' {
			return 0, fmt.Errorf("expected closing parenthesis")
		}
		l.i++
		return v, nil
	}
	if isFormulaDigit(l.s[l.i]) || (l.s[l.i] == '.' && l.i+1 < len(l.s) && isFormulaDigit(l.s[l.i+1])) {
		return readFormulaNumber(l)
	}
	path := readFormulaPath(l)
	if path == "" {
		return 0, fmt.Errorf("expected number or path")
	}
	return toFloat(getByPath(ctx, path)), nil
}

func isFormulaDigit(c byte) bool { return c >= '0' && c <= '9' }

func readFormulaNumber(l *formulaLexer) (float64, error) {
	start := l.i
	seenDot := false
	for l.i < len(l.s) {
		c := l.s[l.i]
		if isFormulaDigit(c) {
			l.i++
			continue
		}
		if c == '.' && !seenDot {
			seenDot = true
			l.i++
			continue
		}
		break
	}
	if start == l.i {
		return 0, fmt.Errorf("invalid number")
	}
	return strconv.ParseFloat(l.s[start:l.i], 64)
}

func readFormulaPath(l *formulaLexer) string {
	start := l.i
	for l.i < len(l.s) {
		c := l.s[l.i]
		if unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '_' || c == '.' {
			l.i++
			continue
		}
		break
	}
	return l.s[start:l.i]
}
