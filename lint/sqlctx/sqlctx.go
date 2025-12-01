// Package sqlctx 检查是否使用了context版本
package sqlctx

import (
	"fmt"
	"go/ast"
	"go/types"
	"maps"
	"sync"

	"golang.org/x/tools/go/analysis"
)

var typesCache sync.Map

const stdContext = "context.Context"

var allowList = map[string]struct{}{
	"os/signal.NotifyContext": {},
}

func report(pass *analysis.Pass, node ast.Node, s string) {
	if _, ok := allowList[s]; ok {
		return
	}
	pass.Report(analysis.Diagnostic{
		Pos:     node.Pos(),
		End:     node.End(),
		Message: fmt.Sprintf("%s should be used instead", s),
	})
}

func getMethods(t types.Type) (map[string]*types.Func, error) {
	if t, ok := typesCache.Load(t); ok {
		return t.(map[string]*types.Func), nil
	}
	var ms map[string]*types.Func
	switch vt := t.(type) {
	case *types.Interface:
		ms = make(map[string]*types.Func, vt.NumMethods())
		for i := 0; i < vt.NumMethods(); i++ {
			ms[vt.Method(i).Name()] = vt.Method(i)
		}
	case *types.Pointer:
		return getMethods(vt.Elem())
	case *types.Alias:
		return getMethods(vt.Underlying())
	case *types.Named:
		ms = make(map[string]*types.Func, vt.NumMethods())
		for i := 0; i < vt.NumMethods(); i++ {
			ms[vt.Method(i).Name()] = vt.Method(i)
		}
		methods, _ := getMethods(vt.Underlying())
		maps.Copy(ms, methods)
	case *types.Struct:
		ms = make(map[string]*types.Func)
		for i := 0; i < vt.NumFields(); i++ {
			if f := vt.Field(i); f.Anonymous() {
				if m, err := getMethods(f.Type()); err == nil {
					for k, v := range m {
						ms[k] = v
					}
				}
			}
		}
	case *types.TypeParam:
		return getMethods(vt.Underlying())
	default:
		return nil, fmt.Errorf("unsupported type %T", t)
	}
	typesCache.Store(t, ms)
	return ms, nil
}

func checkContextVersion(pass *analysis.Pass, fun ast.Expr) {
	if f, ok := pass.TypesInfo.TypeOf(fun).(*types.Signature); !ok {
		return
	} else if f.Params().Len() > 0 && f.Params().At(0).Type().String() == stdContext {
		return
	}
	switch c := fun.(type) {
	case *ast.SelectorExpr:
		t := pass.TypesInfo.TypeOf(c.X)
		var methods map[string]*types.Func
		var err error
		if tt, ok := t.(*types.Basic); ok && tt.Kind() == types.Invalid {
			if t, ok := pass.TypesInfo.ObjectOf(c.Sel).(*types.Func); ok {
				if _, ok = allowList[t.Pkg().Path()+"."+t.Name()]; ok {
					return
				}
				if t.Pkg().Scope().Lookup(t.Name()+"Context") != nil {
					report(pass, c, fmt.Sprintf("%s.%sContext", t.Pkg().Path(), t.Name()))
				}
			}
		} else {
			methods, err = getMethods(t)
			if err != nil {
				report(pass, c, err.Error())
			}
			if _, ok := methods[c.Sel.String()+"Context"]; ok {
				report(pass, c, fmt.Sprintf("%s.%sContext", t.String(), c.Sel.String()))
			}
		}
	case *ast.Ident:
		if t, ok := pass.TypesInfo.ObjectOf(c).(*types.Func); ok {
			if t.Pkg().Scope().Lookup(t.Name()+"Context") != nil {
				report(pass, c, fmt.Sprintf("%s.%sContext", t.Pkg().Path(), t.Name()))
			}
		}
	case *ast.IndexExpr, *ast.ParenExpr, *ast.CallExpr, *ast.FuncLit, *ast.ArrayType:
	case *ast.IndexListExpr:
		checkContextVersion(pass, c.X)
	default:
		report(pass, c, fmt.Sprintf("unsupported call type %T", c))
	}
}

// Analyzer 实现检查是否使用了Context版本的方法
var Analyzer = analysis.Analyzer{
	Name: "sqlctx",
	Doc:  "should use context version of sql function",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				checkContextVersion(pass, call.Fun)
				return true
			})
		}
		return nil, nil
	},
}
