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

// Package renderer contains an implementation of markdown => text/gemini
// renderer for github.com/gomarkdown/markdown.
package renderer

import (
	"fmt"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

var (
	itemIndent = []byte{'\t'}
	itemPrefix = []byte("* ")
)

func (r Renderer) renderFootnotes(w io.Writer, links []ast.Node) (count uint) {
	for _, link := range links {
		if link, ok := link.(*ast.Link); ok && link.Footnote != nil {
			r.link(w, link, true)
			w.Write(lineBreak)
			count++
		}
	}
	return
}

func (r Renderer) renderImages(w io.Writer, links []ast.Node) (count uint) {
	for _, link := range links {
		if link, ok := link.(*ast.Image); ok {
			r.image(w, link, true)
			w.Write(lineBreak)
			count++
		}
	}
	return
}

func (r Renderer) renderLinks(w io.Writer, links []ast.Node) (count uint) {
	for _, link := range links {
		if link, ok := link.(*ast.Link); ok && link.Footnote == nil {
			r.link(w, link, true)
			w.Write(lineBreak)
			count++
		}
	}
	return
}

func (r Renderer) linksList(w io.Writer, links []ast.Node) {
	for _, renderer := range []func(Renderer, io.Writer, []ast.Node) uint{
		Renderer.renderFootnotes,
		Renderer.renderImages,
		Renderer.renderLinks,
	} {
		linksRendered := renderer(r, w, links)
		// ensure breaks between link blocks of the same type
		if linksRendered > 0 {
			w.Write(lineBreak)
		}
	}
}

func (r Renderer) list(w io.Writer, node *ast.List, level int) {
	// the text/gemini spec included with the current Gemini spec does
	// not specify anything about the formatting of lists of level >= 2,
	// as of now this will just render them like in Markdown
	isNumbered := (node.ListFlags & ast.ListTypeOrdered) != 0
	for number, item := range node.Children {
		item, ok := item.(*ast.ListItem)
		if !ok {
			panic("rendering anything but list items is not supported")
		}
		isTerm := (item.ListFlags & ast.ListTypeTerm) == ast.ListTypeTerm
		if l := len(item.Children); l >= 1 {
			// add extra line break to split up definitions
			if isTerm && number > 0 {
				w.Write(lineBreak)
			}
			for i := 0; i < level; i++ {
				w.Write(itemIndent)
			}
			if isNumbered {
				w.Write([]byte(fmt.Sprintf("%d. ", number+1)))
			} else if !isTerm {
				w.Write(itemPrefix)
			}
			r.text(w, item, true)
			w.Write(lineBreak)
			if l >= 2 {
				if list, ok := item.Children[1].(*ast.List); ok {
					r.list(w, list, level+1)
				}
			}
		}
	}
}
