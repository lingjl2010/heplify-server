//nolint:lint
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

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

func main() {
	multichecker.Main(
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
	)
}
