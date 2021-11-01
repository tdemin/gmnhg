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
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

// Based on https://pages.uoregon.edu/ncp/Courses/MathInPlainTextEmail.html
var (
	subOpen  = []byte("_{")
	subClose = []byte("}")
	supOpen  = []byte("^(")
	supClose = []byte(")")
)

func (r Renderer) subscript(w io.Writer, node *ast.Subscript, entering bool) {
	if entering {
		if node := node.AsLeaf(); node != nil {
			w.Write(subOpen)
			w.Write(bytes.ReplaceAll(node.Literal, lineBreak, space))
			w.Write(subClose)
		}
	}
}

func (r Renderer) superscript(w io.Writer, node *ast.Superscript, entering bool) {
	if entering {
		if node := node.AsLeaf(); node != nil {
			w.Write(supOpen)
			w.Write(bytes.ReplaceAll(node.Literal, lineBreak, space))
			w.Write(supClose)
		}
	}
}
