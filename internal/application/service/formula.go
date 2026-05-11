package service

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

// formulaHaltError is the sentinel value used to abort runaway formula execution
// via otto's Interrupt channel. We recover from it inside evalFormula.
var formulaHaltError = errors.New("formula execution timed out")

// numericMemberSegmentRe matches a member-access whose property is a number, e.g. `obj.5`,
// `arr[0].12`, or `foo().3`. The first capture is the character preceding the dot; the
// second is the digit run. We use it to rewrite legacy dotted-path formulas such as
// `event.data.rating.5` into valid JavaScript (`event.data.rating["5"]`) before handing the
// expression to otto. The leading character class deliberately excludes digits so that
// numeric literals like `1.5` are left untouched.
var numericMemberSegmentRe = regexp.MustCompile(`([A-Za-z_$\])])\.(\d+)`)

// formulaTimeout bounds how long a single formula evaluation may run before being
// halted. Formulas are expected to be tiny arithmetic expressions, so this is short.
const formulaTimeout = 200 * time.Millisecond

// evalFormula evaluates a numeric expression using an embedded JavaScript interpreter (otto).
// The expression may reference any top-level keys in ctx (e.g. `metadata.counts_a`,
// `event.data.amount`) just like the previous parser, but it now also supports the full set
// of JavaScript operators (e.g. `Math.min`, ternaries, etc.).
//
// To preserve the previous parser's contract, division by zero (and any other non-finite
// result such as NaN) yields 0 rather than Infinity/NaN.
func evalFormula(expr string, ctx map[string]interface{}) (float64, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, fmt.Errorf("empty formula")
	}
	expr = rewriteNumericMemberSegments(expr)

	vm := otto.New()
	for key, value := range ctx {
		if err := vm.Set(key, value); err != nil {
			return 0, fmt.Errorf("failed to bind %q: %w", key, err)
		}
	}

	value, err := runWithTimeout(vm, expr, formulaTimeout)
	if err != nil {
		return 0, err
	}

	result, err := value.ToFloat()
	if err != nil {
		return 0, err
	}
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, nil
	}
	return result, nil
}

// runWithTimeout executes src on vm, halting it via the Interrupt channel if it
// runs longer than timeout. It recovers from the halt panic that otto raises so
// callers see a normal error instead.
func runWithTimeout(vm *otto.Otto, src string, timeout time.Duration) (value otto.Value, err error) {
	vm.Interrupt = make(chan func(), 1)
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-time.After(timeout):
			vm.Interrupt <- func() { panic(formulaHaltError) }
		case <-done:
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			if r == formulaHaltError {
				err = formulaHaltError
				return
			}
			panic(r)
		}
	}()

	return vm.Run(src)
}

// rewriteNumericMemberSegments rewrites legacy dotted-path segments that contain a numeric
// key (e.g. `event.data.rating.5`) into bracket notation (`event.data.rating["5"]`) so that
// otto's JavaScript parser accepts them. The previous in-house parser allowed numeric path
// segments because it split paths on `.`; this preprocessor preserves that behavior for
// existing rule definitions without changing the rule schema.
//
// The regex is applied repeatedly to handle chained numeric segments like `obj.5.6`, which
// must become `obj["5"]["6"]`.
func rewriteNumericMemberSegments(expr string) string {
	for {
		next := numericMemberSegmentRe.ReplaceAllString(expr, `$1["$2"]`)
		if next == expr {
			return expr
		}
		expr = next
	}
}
