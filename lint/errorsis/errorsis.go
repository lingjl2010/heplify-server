// Package errorsis 检查是否直接使用err比较
package errorsis

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"git.woa.com/tccc/voice_core/lint/lintutils"
)

func isErr(pass *analysis.Pass, expr ast.Expr) bool {
	return pass.TypesInfo.TypeOf(expr).String() == "error"
}

// 判断是否是err变量
// io.EOF 这种虽然也是err变量 但是不应该被检查
// foo.err 这种可能需要被检查
func isErrVariable(pass *analysis.Pass, expr ast.Expr) bool {
	if !isErr(pass, expr) {
		return false
	}
	switch v := expr.(type) {
	case *ast.Ident:
		// 如果变量是个单独的符号，只要不是全局导出的，应该都是err变量
		o := pass.TypesInfo.ObjectOf(v)
		if o.Exported() {
			return false
		}
		// 如果是个本包的全局私有变量，那么也不是err变量
		if o.Parent() == o.Pkg().Scope() {
			return false
		}
		return true
	case *ast.SelectorExpr:
		// x.y 这种就比较复杂 如果是引用别的包的全局变量，那么就不是err变量
		// 如果是引用某个对象的属性，那么就是err变量
		if t, ok := pass.TypesInfo.TypeOf(v.X).(*types.Basic); ok && t.Kind() == types.Invalid {
			// x是包名
			return false
		}
		return true
	case *ast.CallExpr:
		return true
	}
	return false
}

func isNil(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func replaceBinaryOp(n *ast.BinaryExpr, l, r ast.Expr) string {
	x := types.ExprString(l)
	y := types.ExprString(r)
	if n.Op == token.EQL {
		return "errors.Is(" + x + ", " + y + ")"
	}
	return "!errors.Is(" + x + ", " + y + ")"
}

func replaceSwitchStmt(n *ast.SwitchStmt) (fixes []analysis.TextEdit, ok bool) {
	// remove err from switch
	fixes = append(fixes, analysis.TextEdit{
		Pos:     n.Tag.Pos(),
		End:     n.Tag.End(),
		NewText: []byte(""),
	})
	tag := types.ExprString(n.Tag)
	for _, stmt := range n.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok {
			return nil, false
		}
		if cc.List == nil {
			continue
		}
		conditions := make([]string, 0, len(cc.List))
		for _, expr := range cc.List {
			v := types.ExprString(expr)
			conditions = append(conditions, "errors.Is("+tag+", "+v+")")
		}
		fixes = append(fixes, analysis.TextEdit{
			Pos:     cc.List[0].Pos(),
			End:     cc.Colon,
			NewText: []byte(strings.Join(conditions, " || ")),
		})
	}
	return fixes, true
}

// Analyzer 检查是否直接使用
var Analyzer = analysis.Analyzer{
	Name: "errorsis",
	Doc:  "direct comparison of errors",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.BinaryExpr:
					switch n.Op {
					case token.EQL, token.NEQ:
						var l, r ast.Expr
						var lIsErrVariable, rIsErrVariable = isErrVariable(pass, n.X), isErrVariable(pass, n.Y)
						switch {
						case lIsErrVariable && rIsErrVariable:
							pass.Report(analysis.Diagnostic{
								Pos:     n.Pos(),
								End:     n.End(),
								Message: "direct comparison of two complex errors",
							})
							return true
						case lIsErrVariable:
							l = n.X
							r = n.Y
						case rIsErrVariable:
							l = n.Y
							r = n.X
						default:
							return true
						}
						if !isNil(r) { // ignore nil comparison
							report := analysis.Diagnostic{
								Pos:     n.Pos(),
								End:     n.End(),
								Message: "direct comparison of errors",
							}
							replace := replaceBinaryOp(n, l, r)
							report.SuggestedFixes = []analysis.SuggestedFix{{
								Message: "use errors.Is",
								TextEdits: []analysis.TextEdit{{
									Pos:     n.Pos(),
									End:     n.End(),
									NewText: []byte(replace),
								}},
							}}
							pass.Report(report)
							lintutils.AddImport(pass, file, "errors")
						}
					}
				case *ast.SwitchStmt:
					if n.Tag != nil {
						if isErr(pass, n.Tag) {
							report := analysis.Diagnostic{
								Pos:     n.Pos(),
								End:     n.End(),
								Message: "direct comparison of errors",
							}
							replace, ok := replaceSwitchStmt(n)
							if ok {
								report.SuggestedFixes = []analysis.SuggestedFix{{
									Message:   "use errors.Is",
									TextEdits: replace,
								}}
							} else {
								panic("unexpected")
							}
							pass.Report(report)
							lintutils.AddImport(pass, file, "errors")
						}
					}
				}
				return true
			})
		}
		return nil, nil
	},
}
