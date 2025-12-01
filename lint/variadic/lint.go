// Package variadic 检查是否把数组误传给变参函数
package variadic

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
)

func canBeNil(t types.Type) bool {
	switch tt := t.(type) {
	case *types.Pointer, *types.Interface, *types.Map, *types.Chan, *types.Slice, *types.Signature:
		return true
	case *types.Basic:
		switch tt.Kind() {
		case types.UntypedNil, types.UnsafePointer:
			return true
		default:
			return false
		}
	case *types.Named, *types.Alias:
		return canBeNil(tt.Underlying())
	default:
		return false
	}
}

func sliceElemCanBeNil(t types.Type) bool {
	s, ok := t.(*types.Slice)
	if !ok {
		return false
	}
	return canBeNil(s.Elem())
}

func isNil(t ast.Expr) bool {
	ss, ok := t.(*ast.Ident)
	if !ok {
		return false
	}
	return ss.String() == "nil"
}

var allowPkgs = []string{
	"github.com/agiledragon/gomonkey/v2",
	"github.com/golang/mock/gomock",
}

func isAllowCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	t, ok := pass.TypesInfo.Types[call.Fun]
	if !ok {
		return false
	}
	if t.IsBuiltin() {
		return true
	}
	if ff, ok := call.Fun.(*ast.SelectorExpr); ok {
		// 忽略mock的函数
		if x, ok := ff.X.(*ast.Ident); ok {
			if xx, ok := pass.TypesInfo.ObjectOf(x).(*types.PkgName); ok &&
				slices.Contains(allowPkgs, xx.Imported().Path()) {
				return true
			}
		}
		if sig, ok := pass.TypesInfo.TypeOf(ff.Sel).(*types.Signature); ok {
			if sig.Recv() != nil && slices.Contains(allowPkgs, sig.Recv().Pkg().Path()) {
				return true
			}
		}
	}
	return false
}

// Analyzer 实现检查是否把nil误传给变参函数
var Analyzer = analysis.Analyzer{
	Name: "variadic",
	Doc:  "pass nil to the last params of variadic function",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if call.Ellipsis != token.NoPos { // 调用包含了数组展开那说明没问题
					return true
				}
				if len(call.Args) == 0 {
					return true
				}
				f, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
				if !ok {
					return true
				}
				if !f.Variadic() { // 忽略非变参函数
					return true
				}
				if f.Params().Len() == 0 { // 如果没参数也不会误用
					return true
				}
				lastParamIndex := f.Params().Len() - 1
				if !sliceElemCanBeNil(f.Params().At(lastParamIndex).Type()) { // 如果不是any变参 也不会误用
					return true
				}
				if len(call.Args) < f.Params().Len() { // 如果实际参数比参数列表多或少，肯定没误用
					return true
				}
				if !isNil(call.Args[len(call.Args)-1]) {
					return true
				}
				if isAllowCall(pass, call) {
					return true
				}
				pass.Report(analysis.Diagnostic{
					Pos:     call.Args[lastParamIndex].Pos(),
					End:     call.Args[lastParamIndex].End(),
					Message: "the last params is nil, maybe misuse",
				})
				return true
			})
		}
		return nil, nil
	},
}
