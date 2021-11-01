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
	"fmt"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

func (r Renderer) link(w io.Writer, node *ast.Link, entering bool) {
	if entering {
		if node.Footnote != nil {
			fmt.Fprintf(w, "[^%d]: %s", node.NoteID, extractText(node.Footnote))
		} else {
			w.Write(linkPrefix)
			w.Write(node.Destination)
			w.Write(space)
			r.text(w, node, true)
		}
	}
}
