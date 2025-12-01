// Package otelstring 检查是否使用了otel.String
package otelstring

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"git.woa.com/tccc/voice_core/lint/lintutils"
)

const (
	defaultPkgName = "git.woa.com/tccc/voice_core/common/otel"
)

var methodList = []string{
	"Stringer",
	"String",
	"StringValue",
	"StringSlice",
	"StringSliceValue",
}

func checkAttributeString(pass *analysis.Pass, file *ast.File, call *ast.CallExpr) {
	if _, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature); !ok {
		return
	}
	switch c := call.Fun.(type) {
	case *ast.SelectorExpr:
		var method string
		for _, s := range methodList {
			if s == c.Sel.String() {
				method = s
			}
		}
		if method == "" {
			return
		}
		i, ok := c.X.(*ast.Ident)
		if !ok {
			return
		}
		o, ok := pass.TypesInfo.Uses[i]
		if !ok {
			return
		}
		pkg, ok := o.(*types.PkgName)
		if !ok {
			return
		}
		if pkg.Imported().Path() != "go.opentelemetry.io/otel/attribute" {
			return
		}
		// TODO key是否需要检查
		// value为常量时不需要修改
		if _, ok = call.Args[1].(*ast.BasicLit); ok {
			return
		}
		pkgAlias := lintutils.AddImport(pass, file, defaultPkgName)
		var newText string
		if pkgAlias == "" {
			newText = method
		} else {
			newText = fmt.Sprintf("%s.%s", pkgAlias, method)
		}
		pass.Report(analysis.Diagnostic{
			Pos:     call.Pos(),
			Message: "should use otel.String instead of attribute.String",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "use otel.String",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     call.Fun.Pos(),
							End:     call.Fun.End(),
							NewText: []byte(newText),
						},
					},
				},
			},
		})
	case *ast.Ident, *ast.IndexExpr, *ast.ParenExpr, *ast.CallExpr, *ast.FuncLit, *ast.ArrayType, *ast.IndexListExpr:
	default:
		pass.Reportf(call.Pos(), "unsupported call type %T", c)
	}
}

// Analyzer 实现检查是否使用了otel.String方法
var Analyzer = analysis.Analyzer{
	Name: "otelstring",
	Doc:  "should use otel.String instead of attribute.String",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				f, ok := node.(*ast.FuncDecl)
				if ok && pass.Pkg.Path() == defaultPkgName {
					for _, s := range methodList {
						if s == f.Name.String() {
							// 避免递归修改自己
							return false
						}
					}
				}
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				checkAttributeString(pass, file, call)
				return true
			})
		}
		return nil, nil
	},
}
