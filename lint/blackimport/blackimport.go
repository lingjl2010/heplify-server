// Package blackimport 检查是否导入了不该使用的包
package blackimport

import (
	"fmt"
	"go/ast"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

var blacklistMap = map[string]string{
	"git.woa.com/rainbow/golang-sdk/v2/confapi":  "git.woa.com/rainbow/golang-sdk/v2/v3/confapi",
	"git.code.oa.com/rainbow/golang-sdk/confapi": "git.code.oa.com/rainbow/golang-sdk/v3/confapi",
}

func checkImports(pass *analysis.Pass, imports []*ast.ImportSpec) {
	for _, i := range imports {
		path, _ := strconv.Unquote(i.Path.Value)
		if replace, ok := blacklistMap[path]; ok {
			quotedReplace := strconv.Quote(replace)
			message := fmt.Sprintf("should use %s instead of %s", quotedReplace, i.Path.Value)
			pass.Report(analysis.Diagnostic{
				Pos:     i.Path.Pos(),
				End:     i.Path.End(),
				Message: message,
				SuggestedFixes: []analysis.SuggestedFix{{
					TextEdits: []analysis.TextEdit{{
						Pos:     i.Path.Pos(),
						End:     i.Path.End(),
						NewText: []byte(quotedReplace),
					}},
				}},
			})
		}
	}
}

// Analyzer 检查是否导入了不该使用的包
var Analyzer = analysis.Analyzer{
	Name: "blackimport",
	Doc:  "check if import black package",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			checkImports(pass, file.Imports)
		}
		return nil, nil
	},
}
