// Package txmix 检查是否在tx中使用非tx变量
package txmix

import (
	"go/ast"
	"go/types"

	"github.com/samber/lo"
	"golang.org/x/tools/go/analysis"
)

var txTypes = map[string]struct{}{
	"database/sql.Tx":              {},
	"database/sql.Conn":            {},
	"github.com/jmoiron/sqlx.Tx":   {},
	"github.com/jmoiron/sqlx.Conn": {},
}

var dbTypes = map[string]struct{}{
	"database/sql.DB":            {},
	"github.com/jmoiron/sqlx.DB": {},
}

func isType(t types.Type, maps map[string]struct{}) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	_, ok := maps[t.String()]
	return ok
}

func collectFuncCall(pass *analysis.Pass, top ast.Node) (hasTx bool, dbCall []ast.Node) {
	ast.Inspect(top, func(node ast.Node) bool {
		switch call := node.(type) {
		case *ast.FuncLit:
			return false // 不递归检查函数
		case *ast.CallExpr:
			switch c := call.Fun.(type) {
			case *ast.SelectorExpr:
				callType, ok := pass.TypesInfo.TypeOf(c.Sel).(*types.Signature)
				if ok && callType.Recv() != nil {
					if isType(callType.Recv().Type(), txTypes) {
						hasTx = true
					}
					if isType(callType.Recv().Type(), dbTypes) {
						dbCall = append(dbCall, c.Sel)
					}
				}
			default:
				return true
			}
			return true
		default:
			return true
		}
	})
	return
}

func checkFunc(pass *analysis.Pass, ft *ast.FuncType, body *ast.BlockStmt) {
	scope, ok := pass.TypesInfo.Scopes[ft]
	if !ok {
		return
	}

	objects := lo.Map(scope.Names(), func(name string, _ int) types.Object {
		return scope.Lookup(name)
	})
	hasTxInObjects := lo.SomeBy(objects, func(obj types.Object) bool {
		return isType(obj.Type(), txTypes)
	})
	dbInObjects := lo.Filter(objects, func(obj types.Object, _ int) bool {
		return isType(obj.Type(), dbTypes)
	})
	hasTxInCode, dbInCode := collectFuncCall(pass, body)
	if !hasTxInObjects && !hasTxInCode {
		return
	}
	for _, node := range dbInCode {
		pass.Report(analysis.Diagnostic{
			Pos:     node.Pos(),
			End:     node.End(),
			Message: "using db and tx in same scope, maybe bug",
		})
	}
	for _, object := range dbInObjects {
		pass.Report(analysis.Diagnostic{
			Pos:     object.Pos(),
			Message: "using db and tx in same scope, maybe bug",
		})
	}
}

// Analyzer 检查是否在tx中使用非tx变量
var Analyzer = analysis.Analyzer{
	Name: "txmix",
	Doc:  "shouldn't use non-tx variable in tx",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.FuncDecl:
					if n.Body == nil {
						return true
					}
					checkFunc(pass, n.Type, n.Body)
				case *ast.FuncLit:
					checkFunc(pass, n.Type, n.Body)
				}
				return true
			})
		}
		return nil, nil
	},
}
