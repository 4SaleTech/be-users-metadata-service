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
	lexer := &formulaLexer{source: expr, position: 0}
	lexer.skipWhitespace()
	if lexer.position >= len(lexer.source) {
		return 0, fmt.Errorf("empty formula")
	}
	value, err := parseAdditiveExpression(lexer, ctx)
	if err != nil {
		return 0, err
	}
	lexer.skipWhitespace()
	if lexer.position < len(lexer.source) {
		return 0, fmt.Errorf("unexpected trailing input")
	}
	return value, nil
}

type formulaLexer struct {
	source   string
	position int
}

func (l *formulaLexer) skipWhitespace() {
	for l.position < len(l.source) && unicode.IsSpace(rune(l.source[l.position])) {
		l.position++
	}
}

func parseAdditiveExpression(lexer *formulaLexer, ctx map[string]interface{}) (float64, error) {
	accumulator, err := parseMultiplicativeExpression(lexer, ctx)
	if err != nil {
		return 0, err
	}
	for {
		lexer.skipWhitespace()
		if lexer.position >= len(lexer.source) {
			return accumulator, nil
		}
		operator := lexer.source[lexer.position]
		if operator != '+' && operator != '-' {
			return accumulator, nil
		}
		lexer.position++
		nextValue, err := parseMultiplicativeExpression(lexer, ctx)
		if err != nil {
			return 0, err
		}
		if operator == '+' {
			accumulator += nextValue
		} else {
			accumulator -= nextValue
		}
	}
}

func parseMultiplicativeExpression(lexer *formulaLexer, ctx map[string]interface{}) (float64, error) {
	accumulator, err := parsePrimaryExpression(lexer, ctx)
	if err != nil {
		return 0, err
	}
	for {
		lexer.skipWhitespace()
		if lexer.position >= len(lexer.source) {
			return accumulator, nil
		}
		operator := lexer.source[lexer.position]
		if operator != '*' && operator != '/' {
			return accumulator, nil
		}
		lexer.position++
		nextValue, err := parsePrimaryExpression(lexer, ctx)
		if err != nil {
			return 0, err
		}
		if operator == '*' {
			accumulator *= nextValue
		} else if nextValue == 0 {
			accumulator = 0
		} else {
			accumulator /= nextValue
		}
	}
}

func parsePrimaryExpression(lexer *formulaLexer, ctx map[string]interface{}) (float64, error) {
	lexer.skipWhitespace()
	if lexer.position >= len(lexer.source) {
		return 0, fmt.Errorf("unexpected end of formula")
	}
	if lexer.source[lexer.position] == '-' {
		lexer.position++
		value, err := parsePrimaryExpression(lexer, ctx)
		return -value, err
	}
	if lexer.source[lexer.position] == '(' {
		lexer.position++
		value, err := parseAdditiveExpression(lexer, ctx)
		if err != nil {
			return 0, err
		}
		lexer.skipWhitespace()
		if lexer.position >= len(lexer.source) || lexer.source[lexer.position] != ')' {
			return 0, fmt.Errorf("expected closing parenthesis")
		}
		lexer.position++
		return value, nil
	}
	currentChar := lexer.source[lexer.position]
	if isFormulaDigit(currentChar) || (currentChar == '.' && lexer.position+1 < len(lexer.source) && isFormulaDigit(lexer.source[lexer.position+1])) {
		return readNumericLiteral(lexer)
	}
	path := readVariablePath(lexer)
	if path == "" {
		return 0, fmt.Errorf("expected number or path")
	}
	return toFloat(getByPath(ctx, path)), nil
}

func isFormulaDigit(c byte) bool { return c >= '0' && c <= '9' }

func readNumericLiteral(lexer *formulaLexer) (float64, error) {
	start := lexer.position
	hasDecimalPoint := false
	for lexer.position < len(lexer.source) {
		c := lexer.source[lexer.position]
		if isFormulaDigit(c) {
			lexer.position++
			continue
		}
		if c == '.' && !hasDecimalPoint {
			hasDecimalPoint = true
			lexer.position++
			continue
		}
		break
	}
	if start == lexer.position {
		return 0, fmt.Errorf("invalid number")
	}
	return strconv.ParseFloat(lexer.source[start:lexer.position], 64)
}

func readVariablePath(lexer *formulaLexer) string {
	start := lexer.position
	for lexer.position < len(lexer.source) {
		c := lexer.source[lexer.position]
		if unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '_' || c == '.' {
			lexer.position++
			continue
		}
		break
	}
	return lexer.source[start:lexer.position]
}
