// Package lintutils 包含了一些通用的lint函数
package lintutils

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// TODO 有些包的最后一段和包名不同
func pkgName(path string) string {
	segs := strings.Split(path, "/")
	return segs[len(segs)-1]
}

var imported sync.Map

func setImported(file *ast.File, pkg string) bool {
	_, loaded := imported.LoadOrStore(fmt.Sprintf("%p-%s", file, pkg), struct{}{})
	return loaded
}

func locateImport(file *ast.File, newText string) analysis.TextEdit {
	// 第一个import组里
	for _, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT && gd.Lparen != token.NoPos {
			return analysis.TextEdit{
				Pos:     gd.Lparen + 1,
				End:     gd.Lparen + 1,
				NewText: []byte("\n\t" + newText),
			}
		}
	}
	// 第一个import前面
	for _, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT && gd.Lparen == token.NoPos {
			return analysis.TextEdit{
				Pos:     gd.Pos(),
				End:     gd.Pos(),
				NewText: []byte(fmt.Sprintf("import %s\n", newText)),
			}
		}
	}
	return analysis.TextEdit{
		Pos:     file.Name.End(),
		End:     file.Name.End(),
		NewText: []byte(fmt.Sprintf("\nimport %s\n", newText)),
	}
}

func addImport(pass *analysis.Pass, file *ast.File, alias, pkg, importName string) {
	if setImported(file, pkg) {
		return
	}
	var newText string
	if alias == "" {
		newText = fmt.Sprintf("\t\"%s\"\n", importName)
	} else {
		newText = fmt.Sprintf("\t%s \"%s\"\n", alias, pkg)
	}
	edit := locateImport(file, newText)
	pass.Report(analysis.Diagnostic{
		Pos:     file.Imports[0].Pos(),
		Message: fmt.Sprintf("should import package %s", pkg),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message:   fmt.Sprintf("import %s", pkg),
				TextEdits: []analysis.TextEdit{edit},
			},
		},
	})
}

// AddImport 给文件添加import包 返回别名
func AddImport(pass *analysis.Pass, file *ast.File, pkg string) string {
	if pass.Pkg.Path() == pkg {
		return ""
	}
	importName := pkgName(pkg)
	nameMap := map[string]struct{}{}
	for _, i := range file.Imports {
		path := strings.Trim(i.Path.Value, `"`)
		if path == pkg {
			if i.Name != nil {
				return i.Name.Name
			}
			return importName
		}
		if i.Name != nil {
			nameMap[i.Name.Name] = struct{}{}
		} else {
			nameMap[pkgName(path)] = struct{}{}
		}
	}
	if _, ok := nameMap[importName]; !ok {
		addImport(pass, file, "", pkg, importName)
		return importName
	}
	for i := 2; ; i++ {
		name := fmt.Sprintf("%s%d", importName, i)
		if _, ok := nameMap[name]; !ok {
			addImport(pass, file, name, pkg, importName)
			return name
		}
	}
}
