// This file is part of gmnhg.

// gmnhg is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gmnhg is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gmnhg. If not, see <https://www.gnu.org/licenses/>.

// Package gemini provides functions to convert Markdown files to
// Gemtext.
package gemini

import (
	"bytes"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tdemin/gmnhg/internal/renderer"
	renderer2 "github.com/tdemin/gmnhg/internal/renderer2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// Settings is a bitmask for renderer preferences.
type Settings uint

// Has returns true if a flag or a set of flags are all set.
func (s Settings) Has(setting Settings) bool {
	return (s & setting) == setting
}

const (
	// Defaults simply renders the document.
	Defaults Settings = 0
)

var trailing = []byte("\n\n")

func stripTrailingNewlines(content []byte) []byte {
	stripped := content
	for li := bytes.LastIndex(stripped, trailing); li != -1; li = bytes.LastIndex(stripped, trailing) {
		if li != len(stripped)-len(trailing) {
			break
		}
		stripped = stripped[:len(stripped)-1]
	}
	return stripped
}

// RenderMarkdown converts Markdown text to Gemtext using gomarkdown. It
// ignores front matter if any has been provided in the text.
func RenderMarkdown(md []byte, settings Settings) (geminiText []byte, err error) {
	ast := markdown.Parse(md, parser.NewWithExtensions(parser.CommonExtensions|
		parser.NoEmptyLineBeforeBlock|
		parser.Footnotes))
	content := markdown.Render(ast, renderer.NewRenderer())
	return stripTrailingNewlines(content), nil
}

// RenderMarkdown2 converts Markdown text to Gemtext using Goldmark. It
// ignores front matter if any has been provided in the text.
func RenderMarkdown2(md []byte, settings Settings) (geminiText []byte, err error) {
	r := renderer2.NewRenderer()
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.DefinitionList,
		),
		goldmark.WithRenderer(r),
	)
	ast := gm.Parser().Parse(text.NewReader(md))

	buf := bytes.Buffer{}
	if err := gm.Renderer().Render(&buf, md, ast); err != nil {
		return nil, err
	}
	return stripTrailingNewlines(buf.Bytes()), nil
}
