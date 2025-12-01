package blackimport

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestDateBlackImport(t *testing.T) {
	blacklistMap["time"] = "strconv" // testdata不支持外部包 这里用一个标准库包名
	path, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	analysistest.RunWithSuggestedFixes(t, path, &Analyzer)
}
