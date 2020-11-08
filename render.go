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
