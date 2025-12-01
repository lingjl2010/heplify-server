// Package httpctx 检查是否误传递了context.Context
package httpctx

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const stdContext = "context.Context"

// Analyzer 实现检查是否误传递了context.Context
var Analyzer = analysis.Analyzer{
	Name: "httpctx",
	Doc:  "shouldn't pass gin.Context/tgo.Context as context.Context",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				f, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
				if !ok {
					return true
				}
				for i := 0; i < f.Params().Len(); i++ {
					param := f.Params().At(i)
					if param.Type().String() != stdContext {
						return true
					}
					realParam := pass.TypesInfo.TypeOf(call.Args[i])
					switch realParam.String() {
					case stdContext:
					case "*git.code.oa.com/going/going/tgo.Context", "*github.com/gin-gonic/gin.Context":
						pass.Report(analysis.Diagnostic{
							Pos:     call.Args[i].Pos(),
							End:     call.Args[i].End(),
							Message: fmt.Sprintf("shouldn't pass %s as context.Context", realParam.String()),
							SuggestedFixes: []analysis.SuggestedFix{{
								TextEdits: []analysis.TextEdit{{
									Pos:     call.Args[i].End(),
									End:     call.Args[i].End(),
									NewText: []byte(".Request.Context()"),
								}},
							}},
						})
					default:
						pass.Report(analysis.Diagnostic{
							Pos:     call.Args[i].Pos(),
							End:     call.Args[i].End(),
							Message: fmt.Sprintf("shouldn't pass %s as context.Context", realParam.String()),
						})
					}
				}
				return true
			})
		}
		return nil, nil
	},
}
