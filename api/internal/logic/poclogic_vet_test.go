package logic

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPocLogicDoesNotUseLogxErrorWithPrintfDirective(t *testing.T) {
	sourcePath := filepath.Join(".", "poclogic.go")
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read poclogic.go: %v", err)
	}

	badCall := "logx.Error(\"[Nuclei Templates] SDK failed: %v, trying git clone...\", downloadErr)"
	if strings.Contains(string(content), badCall) {
		t.Fatalf("poclogic.go still uses logx.Error with printf directive")
	}
}
