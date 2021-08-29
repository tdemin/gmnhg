package gemini_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gomarkdown/markdown/ast"

	. "github.com/tdemin/gmnhg/internal/gemini"
)

var (
	divider = "============================================================"
)

func TestHr(t *testing.T) {
	node := new(ast.HorizontalRule)
	t.Run("Entering", testRenderNodeStep(node, true, []byte("---\n\n"), ast.GoToNext))
	t.Run("Entered", testRenderNodeStep(node, false, []byte{}, ast.GoToNext))
}

func testRenderNodeStep(node ast.Node, entering bool, expectedGemini []byte, expectedWalkStatus ast.WalkStatus) func(*testing.T) {
	r := NewRenderer()
	w := new(bytes.Buffer)
	return func(t *testing.T) {
		walkStatus := r.RenderNode(w, node, entering)
		if walkStatus != expectedWalkStatus {
			t.Error(fmt.Sprintf("Walk status %T does not match expected value %T!", walkStatus, expectedWalkStatus))
		}
		if !bytes.Equal(w.Bytes(), expectedGemini) {
			t.Error(fmt.Sprintf("Output does not match expected!\n\nActual output:\n\n%s\n%s\n%s\n\nExpected output:\n\n%s\n%s\n%s", divider, w.Bytes(), divider, divider, expectedGemini, divider))
		}
	}
}
