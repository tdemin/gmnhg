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
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gomarkdown/markdown/ast"
	"github.com/olekukonko/tablewriter"
)

var (
	lineBreak          = []byte{'\n'}
	space              = []byte{' '}
	linkPrefix         = []byte("=> ")
	quotePrefix        = []byte("> ")
	itemPrefix         = []byte("* ")
	itemIndent         = []byte{'\t'}
	preformattedToggle = []byte("```\n")
)

var meaningfulCharsRegex = regexp.MustCompile(`\A[\s]+\z`)

const timestampFormat = "2006-01-02 15:04"

// Metadata provides data necessary for proper post rendering.
type Metadata interface {
	Title() string
	Date() time.Time
}

// Renderer implements markdown.Renderer.
type Renderer struct {
	Metadata Metadata
}

// NewRenderer returns a new Renderer.
func NewRenderer() Renderer {
	return Renderer{}
}

// NewRendererWithMetadata returns a new Renderer initialized with post
// metadata.
func NewRendererWithMetadata(m Metadata) Renderer {
	return Renderer{Metadata: m}
}

func (r Renderer) link(w io.Writer, node *ast.Link, entering bool) {
	if entering {
		w.Write(linkPrefix)
		w.Write(node.Destination)
		w.Write(space)
		r.text(w, node)
	}
}

func (r Renderer) image(w io.Writer, node *ast.Image, entering bool) {
	if entering {
		w.Write(linkPrefix)
		w.Write(node.Destination)
		w.Write(space)
		r.text(w, node)
	}
}

func (r Renderer) blockquote(w io.Writer, node *ast.BlockQuote, entering bool) {
	// TODO: Renderer.blockquote: needs support for subnode rendering;
	// ideally to be merged with paragraph
	if entering {
		if node := node.AsContainer(); node != nil {
			for _, child := range node.Children {
				w.Write(quotePrefix)
				// assume children would be paragraphs
				r.text(w, child)
				// double linebreak to ensure Gemini clients don't merge
				// quotes; gomarkdown assumes separate blockquotes are
				// paragraphs of the same blockquote while we don't
				w.Write(lineBreak)
				w.Write(lineBreak)
			}
		}
	}
}

const gemtextHeadingLevelLimit = 3

func (r Renderer) heading(w io.Writer, node *ast.Heading, entering bool) {
	if entering {
		// pad headings with the relevant number of #-s; Gemini spec allows 3 at
		// maximum before the space, therefore add one after 3 and keep padding
		bufLength := node.Level + 1
		spaceNeeded := node.Level > gemtextHeadingLevelLimit
		if spaceNeeded {
			bufLength++
		}
		heading := make([]byte, bufLength)
		heading[len(heading)-1] = ' '
		for i := 0; i < len(heading)-1; i++ {
			heading[i] = '#'
		}
		if spaceNeeded {
			heading[gemtextHeadingLevelLimit] = ' '
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
		if len(children) >= 2 {
			firstChild := children[0]
			_, elementIsText := firstChild.(*ast.Text)
			asLeaf := firstChild.AsLeaf()
			if elementIsText && asLeaf != nil && len(asLeaf.Literal) == 0 {
				children = children[1:]
			}
		}
		linksOnly := func() bool {
			for _, child := range children {
				// TODO: simplify
				if _, ok := child.(*ast.Link); ok {
					continue
				}
				if _, ok := child.(*ast.Image); ok {
					continue
				}
				if child, ok := child.(*ast.Text); ok {
					// any meaningful text?
					if meaningfulCharsRegex.Find(child.Literal) == nil {
						return false
					}
					continue
				}
				return false
			}
			return true
		}()
		noNewLine = linksOnly
		for _, child := range children {
			// only render links text in the paragraph if they're
			// combined with some other text on page
			switch child := child.(type) {
			case *ast.Link, *ast.Image:
				if !linksOnly {
					r.text(w, child)
				}
				linkStack = append(linkStack, child)
			case *ast.Text, *ast.Code, *ast.Emph, *ast.Strong, *ast.Del:
				// the condition prevents text blocks consisting only of
				// line breaks and spaces and such from rendering
				if !linksOnly {
					r.text(w, child)
				}
			}
		}
		if !linksOnly {
			w.Write(lineBreak)
		}
		// render a links block after paragraph
		if len(linkStack) > 0 {
			if !linksOnly {
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
	w.Write(preformattedToggle)
	w.Write(node.Literal)
	w.Write(preformattedToggle)
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
				r.text(w, para)
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

func (r Renderer) text(w io.Writer, node ast.Node) {
	if node := node.AsLeaf(); node != nil {
		// replace all newlines in text with spaces, allowing for soft
		// wrapping; this is recommended as per Gemini spec p. 5.4.1
		w.Write([]byte(strings.ReplaceAll(string(node.Literal), "\n", " ")))
		return
	}
	if node := node.AsContainer(); node != nil {
		for _, child := range node.Children {
			r.text(w, child)
		}
	}
}

func extractText(node ast.Node) string {
	if node := node.AsLeaf(); node != nil {
		return strings.ReplaceAll(string(node.Literal), "\n", " ")
	}
	if node := node.AsContainer(); node != nil {
		b := strings.Builder{}
		for _, child := range node.Children {
			b.WriteString(extractText(child))
		}
		return b.String()
	}
	panic("encountered a non-leaf & non-container node")
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
	}
}

// RenderNode implements Renderer.RenderNode().
func (r Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	// despite most of the subroutines here accepting entering, most of
	// them don't really need an extra pass
	noNewLine := true
	switch node := node.(type) {
	case *ast.BlockQuote:
		r.blockquote(w, node, entering)
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
	case *ast.Table:
		r.table(w, node, entering)
		noNewLine = false
	}
	if !noNewLine && !entering {
		w.Write(lineBreak)
	}
	return ast.GoToNext
}

// RenderHeader implements Renderer.RenderHeader(). It renders metadata
// at the top of the post if any has been provided.
func (r Renderer) RenderHeader(w io.Writer, node ast.Node) {
	if r.Metadata != nil {
		// TODO: Renderer.RenderHeader: check whether date is mandatory
		// in Hugo
		w.Write([]byte(fmt.Sprintf("# %s\n\n%s\n\n", r.Metadata.Title(), r.Metadata.Date().Format(timestampFormat))))
	}
}

// RenderFooter implements Renderer.RenderFooter().
func (r Renderer) RenderFooter(w io.Writer, node ast.Node) {}
