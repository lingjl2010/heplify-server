// Package passpointer 检查是否传递了指针给常见函数
// 参考代码 golang.org/x/tools/go/analysis/passes/unmarshal
package passpointer

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

func not[T any](f func(T) bool) func(T) bool {
	return func(s T) bool {
		return !f(s)
	}
}

// fix match
func f[T comparable](matches ...T) func(T) bool {
	return func(s T) bool {
		return slices.Contains(matches, s)
	}
}

// any match
func a[T any]() func(T) bool {
	return func(T) bool {
		return true
	}
}

type rule struct {
	PkgMatcher func(string) bool
	FnMatcher  func(string) bool
	ArgMatcher func(int) bool
}

func (r rule) match(pkg, fn string) bool {
	return r.PkgMatcher(pkg) && r.FnMatcher(fn)
}

var rules = []rule{
	// govet已经有标准库的unmarshal检查，这里不再检查
	// {a[string](), f("Unmarshal"), f(1)},
	{f("fmt"), f("Scan", "Scanln"), a[int]()},
	{f("fmt"), f("Scanf", "Fscan", "Fscanln", "Sscan", "Sscanln"), not(f[int](0))},
	{f("fmt"), f("Sscanf", "Fscanf"), not(f[int](0, 1))},
	{f("database/sql"), f("Scan"), a[int]()},
	{f("github.com/jmoiron/sqlx"), f("Get", "Select"), f(0)},
	{f("github.com/jmoiron/sqlx"), f("GetContext", "SelectContext"), f(1)},
	{f("github.com/jmoiron/sqlx"), f("Scan", "StructScan"), a[int]()},
}

// Analyzer 实现检查是否传递了指针给常见函数
var Analyzer = analysis.Analyzer{
	Name:     "passpointer",
	Doc:      "should pass pointer values to some functions",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) // nolint:errcheck
	inspect.Preorder([]ast.Node{
		(*ast.CallExpr)(nil),
	}, func(n ast.Node) {
		call := n.(*ast.CallExpr) // nolint:errcheck
		fn := typeutil.StaticCallee(pass.TypesInfo, call)
		if fn == nil {
			return // not a static call
		}

		for _, rule := range rules {
			if !rule.match(fn.Pkg().Path(), fn.Name()) {
				continue
			}
			var badIndexes []int
			for argidx := range call.Args {
				if !rule.ArgMatcher(argidx) {
					continue
				}
				t := pass.TypesInfo.Types[call.Args[argidx]].Type
				switch t.Underlying().(type) {
				case *types.Pointer, *types.Interface, *types.TypeParam:
					continue
				case *types.Slice: // 忽略变长参数
					if call.Ellipsis != token.NoPos && argidx == len(call.Args)-1 {
						continue
					}
				}
				badIndexes = append(badIndexes, argidx)
			}
			if len(badIndexes) > 0 {
				pass.Reportf(call.Lparen, "call of %s(%v) passes non-pointer", fn.Name(), badIndexes)
			}
		}

	})
	return nil, nil
}
