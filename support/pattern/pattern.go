package pattern

import (
	"fmt"
	"strings"
)

type Pattern struct {
	exprs []expr
}
type Patterns []*Pattern

func CompilePatterns(patterns []string) Patterns {
	list := make([]*Pattern, 0, len(patterns))
	for _, p := range patterns {
		ptn := CompilePattern(p)
		list = append(list, ptn)
	}
	return list
}

func CompilePattern(s string) *Pattern {
	segments := splitPath(s)
	exprs := make([]expr, 0, len(segments))
	for _, seg := range segments {
		expr := compileExpr(seg)
		exprs = append(exprs, expr)
	}
	return &Pattern{exprs: exprs}
}

type kind int

const (
	kind_plain_str = iota
	kind_star      // *
)

type element struct {
	kind  kind
	runes []rune
}

type elements []element

func (c *Pattern) MatchPrefix(path string) bool {
	return matchSegsPrefix(c.exprs, splitPath(path))
}

// match exact
func (c *Pattern) Match(path string) bool {
	return matchSegsFull(c.exprs, splitPath(path))
}

func (c *Pattern) matchPrefixPaths(paths []string) bool {
	return matchSegsPrefix(c.exprs, paths)
}

func (c Patterns) MatchAnyPrefix(path string) bool {
	paths := splitPath(path)
	return matchAnyPatterns(c, paths)
}

func (c Patterns) MatchAny(path string) bool {
	paths := splitPath(path)
	for _, pattern := range c {
		if matchSegsFull(pattern.exprs, paths) {
			return true
		}
	}
	return false
}

func (c Patterns) matchAnyPrefixPaths(paths []string) bool {
	return matchAnyPatterns(c, paths)
}
func matchAnyPatterns(patterns []*Pattern, paths []string) bool {
	for _, pattern := range patterns {
		if pattern.matchPrefixPaths(paths) {
			return true
		}
	}
	return false
}

type expr struct {
	doubleStar bool
	elements   elements
}

func compileExpr(s string) expr {
	if s == "" {
		return expr{}
	}
	if s == "**" {
		return expr{doubleStar: true}
	}
	runes := []rune(s)

	elems := make(elements, 0)

	lastIdx := 0
	for i, ch := range runes {
		if ch != '*' {
			continue
		}
		if i > lastIdx {
			elems = append(elems, element{kind: kind_plain_str, runes: runes[lastIdx:i]})
		}
		lastIdx = i + 1
		if i > 0 && runes[i-1] == '*' {
			continue
		}
		elems = append(elems, element{kind: kind_star})
	}
	if lastIdx < len(runes) {
		elems = append(elems, element{kind: kind_plain_str, runes: runes[lastIdx:]})
	}
	return expr{elements: elems}
}

func splitPath(path string) []string {
	segments := strings.Split(path, "/")
	filtered := make([]string, 0, len(segments))
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		filtered = append(filtered, seg)
	}
	return filtered
}

func matchSegsPrefix(exprs []expr, segments []string) bool {
	return doMatch(exprs, segments, true)
}

func matchSegsFull(exprs []expr, segments []string) bool {
	return doMatch(exprs, segments, false)
}

// f(L,j,segs,i) = if L[j] double star: f(L,j,segs,i+1) or f(L,j+1,segs,i); else if L[j] matches segs[i],f(L,j+1,segs,i+1)
// if j>=L.length: if segments empty
func doMatch(exprs []expr, segments []string, prefix bool) bool {
	if len(exprs) == 0 {
		if prefix {
			return true
		}
		return len(segments) == 0
	}
	expr := exprs[0]
	if expr.doubleStar {
		if len(segments) > 0 && doMatch(exprs, segments[1:], prefix) {
			return true
		}
		return doMatch(exprs[1:], segments, prefix)
	}
	if len(segments) == 0 {
		return expr.matchEmpty()
	}

	if !expr.matchNoDoubleStar(segments[0], prefix) {
		return false
	}
	return doMatch(exprs[1:], segments[1:], prefix)
}

func (c expr) matchNoDoubleStar(name string, prefix bool) bool {
	return c.matchRunesFrom(0, prefix, []rune(name))
}

// f(L, i, runes,j ) -> if L[i]==="*", f(L,i+1, runes,j) or f(L,i,runes,j+1)
func (c expr) matchRunesFrom(i int, prefix bool, runes []rune) bool {
	if i >= len(c.elements) {
		return len(runes) == 0
	}
	part := c.elements[i]
	switch part.kind {
	case kind_star:
		if c.matchRunesFrom(i+1, prefix, runes) {
			return true
		}
		if len(runes) > 0 {
			return c.matchRunesFrom(i, prefix, runes[1:])
		}
		return false
	case kind_plain_str:
		n := len(part.runes)
		if !prefix {
			if n != len(runes) {
				return false
			}
		} else {
			if n > len(runes) {
				return false
			}
		}
		for j := 0; j < n; j++ {
			if runes[j] != part.runes[j] {
				return false
			}
		}
		return true
	default:
		panic(fmt.Errorf("unknown expr kind: %v", part.kind))
	}
}

func (c expr) matchEmpty() bool {
	// all are just stars
	for _, part := range c.elements {
		if part.kind != kind_star {
			return false
		}
	}
	return true
}
