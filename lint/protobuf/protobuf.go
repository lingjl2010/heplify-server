// Package protobuf 检查是否正确使用了proto.Unmarshal
package protobuf

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func checkProtoUnmarshal(pass *analysis.Pass, call *ast.CallExpr) {
	switch c := call.Fun.(type) {
	case *ast.SelectorExpr:
		if c.Sel.String() != "Unmarshal" {
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
		if pkg.Imported().Path() != "github.com/golang/protobuf/proto" &&
			pkg.Imported().Path() != "google.golang.org/protobuf/proto" {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos:     call.Fun.Pos(),
			End:     call.Fun.End(),
			Message: "shouldn't use proto.Unmarshal directly",
		})
	case *ast.IndexExpr, *ast.ParenExpr, *ast.CallExpr, *ast.Ident, *ast.FuncLit, *ast.ArrayType, *ast.IndexListExpr:
	default:
		pass.Reportf(call.Pos(), "unsupported call type %T", c)
	}
}

// Analyzer 实现检查误用proto.Unmarshal
var Analyzer = analysis.Analyzer{
	Name: "protobuf",
	Doc:  "shouldn't use proto.Unmarshal directly",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				checkProtoUnmarshal(pass, call)
				return true
			})
		}
		return nil, nil
	},
}
