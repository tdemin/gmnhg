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
	"github.com/olekukonko/tablewriter"
)

func extractText(node ast.Node) string {
	return string(textWithNewlineReplacement(node, space, true))
}

func (r Renderer) tableHead(t *tablewriter.Table, node *ast.TableHeader) {
	if node := node.AsContainer(); node != nil {
		// should always have a single row consisting of at least one
		// cell but worth checking nonetheless; tablewriter only
		// supports a single header row as of now therefore ignore
		// second row and the rest
		if len(node.Children) > 0 {
			if row := node.Children[0].AsContainer(); row != nil {
				cells := make([]string, len(row.Children))
				for i, cell := range row.Children {
					cells[i] = extractText(cell)
				}
				t.SetHeader(cells)
			}
		}
	}
}

func (r Renderer) tableBody(t *tablewriter.Table, node *ast.TableBody) {
	if node := node.AsContainer(); node != nil {
		for _, row := range node.Children {
			if row := row.AsContainer(); row != nil {
				cells := make([]string, len(row.Children))
				for i, cell := range row.Children {
					cells[i] = extractText(cell)
				}
				t.Append(cells)
			}
		}
	}
}

func (r Renderer) table(w io.Writer, node *ast.Table, entering bool) {
	if entering {
		w.Write(preformattedToggle)
		w.Write(lineBreak)
		// gomarkdown appears to only parse headings consisting of a
		// single line and always have a TableBody preceded by a single
		// TableHeader but we're better off not relying on it
		t := tablewriter.NewWriter(w)
		t.SetAutoFormatHeaders(false) // TODO: tablewriter options should probably be configurable
		if node := node.AsContainer(); node != nil {
			for _, child := range node.Children {
				switch child := child.(type) {
				case *ast.TableHeader:
					r.tableHead(t, child)
				case *ast.TableBody:
					r.tableBody(t, child)
				}
			}
		}
		t.Render()
	} else {
		w.Write(preformattedToggle)
		w.Write(lineBreak)
	}
}
