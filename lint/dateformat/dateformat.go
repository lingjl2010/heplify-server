// Package dateformat 检查应该优先使用time的格式
package dateformat

import (
	"fmt"
	"go/ast"
	"time"

	"golang.org/x/tools/go/analysis"
)

var blacklist = map[string]string{
	time.Layout:      "time.Layout",
	time.ANSIC:       "time.ANSIC",
	time.UnixDate:    "time.UnixDate",
	time.RubyDate:    "time.RubyDate",
	time.RFC822:      "time.RFC822",
	time.RFC822Z:     "time.RFC822Z",
	time.RFC850:      "time.RFC850",
	time.RFC1123:     "time.RFC1123",
	time.RFC1123Z:    "time.RFC1123Z",
	time.RFC3339:     "time.RFC3339",
	time.RFC3339Nano: "time.RFC3339Nano",
	time.Kitchen:     "time.Kitchen",
	time.Stamp:       "time.Stamp",
	time.StampMilli:  "time.StampMilli",
	time.StampMicro:  "time.StampMicro",
	time.StampNano:   "time.StampNano",
	time.DateTime:    "time.DateTime",
	time.DateOnly:    "time.DateOnly",
	time.TimeOnly:    "time.TimeOnly",
}

var blacklistMap map[string]string

func init() {
	blacklistMap = make(map[string]string, len(blacklist)*2)
	for key, replace := range blacklist {
		blacklistMap[`"`+key+`"`] = replace
		blacklistMap["`"+key+"`"] = replace
	}
}

func checkText(pass *analysis.Pass, n *ast.BasicLit) {
	if replace, ok := blacklistMap[n.Value]; ok {
		message := fmt.Sprintf("should use %s instead of %s", replace, n.Value)
		pass.Report(analysis.Diagnostic{
			Pos:     n.Pos(),
			End:     n.End(),
			Message: message,
			SuggestedFixes: []analysis.SuggestedFix{{
				TextEdits: []analysis.TextEdit{{
					Pos:     n.Pos(),
					End:     n.End(),
					NewText: []byte(replace),
				}},
			}},
		})
	}
}

// Analyzer 实现检查是否误传递了context.Context
var Analyzer = analysis.Analyzer{
	Name: "dateformat",
	Doc:  "check if use time const layout",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.BasicLit:
					checkText(pass, n)
				default:
					return true
				}
				return true

			})
		}
		return nil, nil
	},
}
