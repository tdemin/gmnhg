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

package renderer

import (
	"io"

	"github.com/gomarkdown/markdown/ast"
)

var (
	quoteBrPrefix = []byte("\n> ")
	quotePrefix   = []byte("> ")
)

func (r Renderer) blockquote(w io.Writer, node *ast.BlockQuote, entering bool) {
	if entering {
		if node := node.AsContainer(); node != nil {
			for _, child := range node.Children {
				w.Write(quotePrefix)
				r.blockquoteText(w, child)
				// double linebreak to ensure Gemini clients don't merge
				// quotes; gomarkdown assumes separate blockquotes are
				// paragraphs of the same blockquote while we don't
				w.Write(lineBreak)
				w.Write(lineBreak)
			}
		}
	}
}

func (r Renderer) blockquoteText(w io.Writer, node ast.Node) {
	w.Write(textWithNewlineReplacement(node, quoteBrPrefix, true))
}
