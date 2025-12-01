// Package dirtyword 检查是否使用了不该使用的词汇
package dirtyword

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type matcher interface {
	match(s string) bool
	keyword() string
}

type stringMatcher string

func (s stringMatcher) match(text string) bool {
	return strings.Contains(text, string(s))
}

func (s stringMatcher) keyword() string {
	return string(s)
}

type regexpMatcher regexp.Regexp

func (r *regexpMatcher) match(text string) bool {
	return (*regexp.Regexp)(r).MatchString(text)
}

func (r *regexpMatcher) keyword() string {
	return (*regexp.Regexp)(r).String()
}

// nolint:lint
var badWords = []matcher{
	stringMatcher("坐席"),
	stringMatcher("帐号"),
	(*regexpMatcher)(regexp.MustCompile(`(?i)replace\s+into`)),
	(*regexpMatcher)(regexp.MustCompile(`(?i)insert\s+ignore\s+into`)),
}

func checkText(pass *analysis.Pass, text string, pos token.Pos) {
	for _, word := range badWords {
		if word.match(text) {
			pass.Report(analysis.Diagnostic{
				Pos:     pos,
				Message: fmt.Sprintf("shouldn't use %s", word.keyword()),
			})
		}
	}
}

// Analyzer 实现检查是否误传递了context.Context
var Analyzer = analysis.Analyzer{
	Name: "dirtyword",
	Doc:  "check if use dirty word",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			for _, commentGroups := range file.Comments {
				for _, comment := range commentGroups.List {
					checkText(pass, comment.Text, comment.Pos())
				}
			}
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.BasicLit:
					checkText(pass, n.Value, n.Pos())
				case *ast.Ident:
					checkText(pass, n.Name, n.Pos())
				default:
					return true
				}
				return true

			})
		}
		return nil, nil
	},
}
