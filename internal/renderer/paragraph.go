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

func isLinksOnlyParagraph(node *ast.Paragraph) bool {
	for _, child := range node.Children {
		switch child := child.(type) {
		case *ast.Text:
			if emptyLineRegex.Find(child.Literal) != nil {
				continue
			}
		case *ast.Link, *ast.Image:
			continue
		}
		return false
	}
	return true
}

func (r Renderer) paragraph(w io.Writer, node *ast.Paragraph, entering bool) (noNewLine bool) {
	linksOnly := isLinksOnlyParagraph(node)
	noNewLine = linksOnly
	if entering {
		children := node.Children
		// current version of gomarkdown/markdown finds an empty
		// *ast.Text element before links/images, breaking the heuristic
		if len(children) >= 2 {
			firstChild, elementIsText := children[0].(*ast.Text)
			if elementIsText && len(firstChild.Literal) == 0 {
				children = children[1:]
			}
		}
		if !linksOnly {
			for _, child := range children {
				// only render links text in the paragraph if they're
				// combined with some other text on page
				switch child := child.(type) {
				case *ast.Text, *ast.Emph, *ast.Strong, *ast.Del, *ast.Link, *ast.Image:
					r.text(w, child, true)
				case *ast.Code:
					r.text(w, child, false)
				case *ast.Hardbreak:
					w.Write(lineBreak)
				case *ast.HTMLSpan:
					if isHardBreak(child.AsLeaf().Literal) {
						w.Write(lineBreak)
					}
				case *ast.Subscript:
					r.subscript(w, child, true)
				case *ast.Superscript:
					r.superscript(w, child, true)
				}
			}
			w.Write(lineBreak)
		}
	}
	return
}
