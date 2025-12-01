// Package module golangci-lint新的module插件
package module

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"git.woa.com/tccc/voice_core/lint/blackimport"
	"git.woa.com/tccc/voice_core/lint/dateformat"
	"git.woa.com/tccc/voice_core/lint/dirtyword"
	"git.woa.com/tccc/voice_core/lint/errorsis"
	"git.woa.com/tccc/voice_core/lint/executable"
	"git.woa.com/tccc/voice_core/lint/httpctx"
	"git.woa.com/tccc/voice_core/lint/otelstring"
	"git.woa.com/tccc/voice_core/lint/passpointer"
	"git.woa.com/tccc/voice_core/lint/protobuf"
	"git.woa.com/tccc/voice_core/lint/sqlctx"
	"git.woa.com/tccc/voice_core/lint/txmix"
	"git.woa.com/tccc/voice_core/lint/variadic"
)

func init() {
	register.Plugin("lint", New)
}

type plugin struct{}

func (p plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		&httpctx.Analyzer,
		&executable.Analyzer,
		&sqlctx.Analyzer,
		&protobuf.Analyzer,
		&otelstring.Analyzer,
		&variadic.Analyzer,
		&dirtyword.Analyzer,
		&dateformat.Analyzer,
		&blackimport.Analyzer,
		&txmix.Analyzer,
		&passpointer.Analyzer,
		&errorsis.Analyzer,
	}, nil
}

func (p plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// New 创建所有插件
func New(conf any) (register.LinterPlugin, error) {
	return plugin{}, nil
}
