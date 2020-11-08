package gemini

import (
	"git.tdem.in/tdemin/gmnhg/internal/gemini"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

// RenderMarkdown converts Markdown text to text/gemini using gomarkdown.
//
// gomarkdown doesn't return any errors, nor does this function.
func RenderMarkdown(md []byte) (geminiText []byte) {
	ast := markdown.Parse(md, parser.NewWithExtensions(parser.CommonExtensions))
	geminiContent := markdown.Render(ast, gemini.NewRenderer())
	return geminiContent
}
