// Package executable 检查是否包含了executable包
package executable

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	executable   = `"git.woa.com/tccc/voice_core/common/executable"`
	executableFx = `"git.woa.com/tccc/voice_core/common/executablefx"`
	main         = "main"
)

func isMain(f *ast.File) bool {
	if f.Name.String() != main {
		return false
	}
	for _, decl := range f.Decls {
		if fun, ok := decl.(*ast.FuncDecl); ok {
			if fun.Name.String() == main {
				return true
			}
		}
	}
	return false
}

func checkExecutable(f *ast.File, pass *analysis.Pass) {
	for _, imp := range f.Imports {
		if imp.Path.Value == executable || imp.Path.Value == executableFx {
			if !isMain(f) {
				pass.Reportf(imp.Pos(), "shouldn't import %s package in non-main file", executable)
			}
			return
		}
	}
	if isMain(f) {
		pass.Reportf(f.Name.Pos(), "should import %s package in main file", executable)
	}
}

// Analyzer 实现检查可执行文件是否满足需求
var Analyzer = analysis.Analyzer{
	Name: "executable",
	Doc:  "should use executable correctly",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			checkExecutable(file, pass)
		}
		return nil, nil
	},
}
