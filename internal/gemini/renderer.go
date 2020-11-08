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

// Package gemini contains an implementation of markdown => text/gemini
// renderer for github.com/gomarkdown/markdown.
package gemini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

var (
	lineBreak   = []byte{'\n'}
	space       = []byte{' '}
	linkPrefix  = []byte("=> ")
	quotePrefix = []byte("> ")
	itemPrefix  = []byte("* ")
	itemIndent  = []byte{'\t'}
)

// Renderer implements markdown.Renderer.
type Renderer struct{}

// NewRenderer returns a new Renderer.
func NewRenderer() Renderer {
	return Renderer{}
}

func (r Renderer) link(w io.Writer, node *ast.Link, entering bool) {
	if entering {
		w.Write(linkPrefix)
		w.Write(node.Destination)
		for _, child := range node.Children {
			if l := child.AsLeaf(); l != nil {
				w.Write(space)
				w.Write(l.Literal)
			}
		}
	}
}

func (r Renderer) linkText(w io.Writer, node *ast.Link) {
	for _, text := range node.Children {
		// TODO: Renderer.linkText: link can contain subblocks
		if l := text.AsLeaf(); l != nil {
			w.Write(l.Literal)
		}
	}
}

func (r Renderer) imageText(w io.Writer, node *ast.Image) {
	for _, text := range node.Children {
		// TODO: Renderer.imageText: link can contain subblocks
		if l := text.AsLeaf(); l != nil {
			w.Write(l.Literal)
		}
	}
}

func (r Renderer) image(w io.Writer, node *ast.Image, entering bool) {
	if entering {
		w.Write(linkPrefix)
		w.Write(node.Destination)
		for _, sub := range node.Container.Children {
			if l := sub.AsLeaf(); l != nil {
				// TODO: Renderer.image: Markdown technically allows for
				// links inside image titles, yet to think out how to
				// render that :thinking:
				w.Write(space)
				w.Write(l.Literal)
			}
		}
	}
}

func (r Renderer) blockquote(w io.Writer, node *ast.BlockQuote, entering bool) {
	// TODO: Renderer.blockquote: needs support for subnode rendering;
	// ideally to be merged with paragraph
	if entering {
		if para, ok := node.Children[0].(*ast.Paragraph); ok {
			for _, subnode := range para.Children {
				if l := subnode.AsLeaf(); l != nil {
					reader := bufio.NewScanner(bytes.NewBuffer(l.Literal))
					for reader.Scan() {
						w.Write(quotePrefix)
						w.Write(reader.Bytes())
						w.Write(lineBreak)
					}
				}
			}
		}
	}
}

func (r Renderer) heading(w io.Writer, node *ast.Heading, entering bool) {
	if entering {
		// pad headings with the relevant number of #-s
		heading := make([]byte, node.Level+1)
		heading[len(heading)-1] = ' '
		for i := 0; i < len(heading)-1; i++ {
			heading[i] = '#'
		}
		w.Write(heading)
		for _, text := range node.Children {
			w.Write(text.AsLeaf().Literal)
		}
	} else {
		w.Write(lineBreak)
	}
}

func (r Renderer) paragraph(w io.Writer, node *ast.Paragraph, entering bool) (noNewLine bool) {
	if entering {
		children := node.Children
		linkStack := make([]ast.Node, 0, len(children))
		// current version of gomarkdown/markdown finds an empty
		// *ast.Text element before links/images, breaking the heuristic
		onlyElementWithGoMarkdownFix := func() bool {
			if len(node.Children) > 1 {
				firstChild := node.Children[0]
				_, elementIsText := firstChild.(*ast.Text)
				asLeaf := firstChild.AsLeaf()
				if elementIsText && asLeaf != nil && len(asLeaf.Literal) == 0 {
					children = children[1:]
					return true
				}
			}
			return false
		}()
		onlyElement := len(children) == 1 || onlyElementWithGoMarkdownFix
		onlyElementIsLink := func() bool {
			if len(children) >= 1 {
				if _, ok := children[0].(*ast.Link); ok {
					return true
				}
				if _, ok := children[0].(*ast.Image); ok {
					return true
				}
			}
			return false
		}()
		noNewLine = onlyElementIsLink
		for _, child := range children {
			// only render links text in the paragraph if they're
			// combined with some other text on page
			if link, ok := child.(*ast.Link); ok {
				if !onlyElement {
					r.linkText(w, link)
				}
				linkStack = append(linkStack, link)
			}
			if image, ok := child.(*ast.Image); ok {
				if !onlyElement {
					r.imageText(w, image)
				}
				linkStack = append(linkStack, image)
			}
			if text, ok := child.(*ast.Text); ok {
				r.text(w, text)
			}
		}
		if !onlyElementIsLink {
			w.Write(lineBreak)
		}
		// render a links block after paragraph
		if len(linkStack) > 0 {
			if !onlyElementIsLink {
				w.Write(lineBreak)
			}
			for _, link := range linkStack {
				if link, ok := link.(*ast.Link); ok {
					r.link(w, link, true)
				}
				if image, ok := link.(*ast.Image); ok {
					r.image(w, image, true)
				}
				w.Write(lineBreak)
			}
		}
	}
	return
}

func (r Renderer) code(w io.Writer, node *ast.CodeBlock) {
	w.Write([]byte("```\n"))
	w.Write(node.Literal)
	w.Write([]byte("```\n"))
}

func (r Renderer) list(w io.Writer, node *ast.List, level int) {
	// the text/gemini spec included with the current Gemini spec does
	// not specify anything about the formatting of lists of level >= 2,
	// as of now this will just render them like in Markdown
	isNumbered := (node.ListFlags & ast.ListTypeOrdered) == ast.ListTypeOrdered
	for number, item := range node.Children {
		item, ok := item.(*ast.ListItem)
		if !ok {
			panic("rendering anything but list items is not supported")
		}
		// this assumes github.com/gomarkdown/markdown can only produce
		// list items that contain a child paragraph and possibly
		// another list; this might not be true but I can hardly imagine
		// a list item that contains anything else
		if l := len(item.Children); l >= 1 {
			for i := 0; i < level; i++ {
				w.Write(itemIndent)
			}
			if isNumbered {
				w.Write([]byte(fmt.Sprintf("%d. ", number+1)))
			} else {
				w.Write(itemPrefix)
			}
			para, ok := item.Children[0].(*ast.Paragraph)
			if ok {
				text, ok := para.Children[0].(*ast.Text)
				if ok {
					r.text(w, text)
				}
			}
			w.Write(lineBreak)
			if l >= 2 {
				if list, ok := item.Children[1].(*ast.List); ok {
					r.list(w, list, level+1)
				}
			}
		}
	}
}

func (r Renderer) text(w io.Writer, node *ast.Text) {
	w.Write(node.Literal)
}

// RenderNode implements Renderer.RenderNode().
func (r Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	// despite most of the subroutines here accepting entering, most of
	// them don't really need an extra pass
	noNewLine := true
	switch node := node.(type) {
	case *ast.BlockQuote:
		r.blockquote(w, node, entering)
		noNewLine = false
	case *ast.Heading:
		r.heading(w, node, entering)
		noNewLine = false
	case *ast.Paragraph:
		// blockquote wraps paragraphs which makes for an extra render
		_, parentIsBlockQuote := node.Parent.(*ast.BlockQuote)
		_, parentIsListItem := node.Parent.(*ast.ListItem)
		if !parentIsBlockQuote && !parentIsListItem {
			noNewLine = r.paragraph(w, node, entering)
		}
	case *ast.CodeBlock:
		r.code(w, node)
		// code block is not considered a wrapping element
		w.Write(lineBreak)
	case *ast.List:
		// lists of level >= 2 are rendered recursively along with the
		// first level; the list is a container
		if _, parentIsDocument := node.Parent.(*ast.Document); parentIsDocument && !entering {
			r.list(w, node, 0)
			noNewLine = false
		}
	}
	if !noNewLine && !entering {
		w.Write(lineBreak)
	}
	return ast.GoToNext
}

// RenderHeader implements Renderer.RenderHeader().
func (r Renderer) RenderHeader(w io.Writer, node ast.Node) {
	// likely doesn't need any code
}

// RenderFooter implements Renderer.RenderFooter().
func (r Renderer) RenderFooter(w io.Writer, node ast.Node) {
	// likely doesn't need any code either
}
