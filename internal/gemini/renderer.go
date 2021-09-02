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
	quoteBrPrefix      = []byte("\n> ")
	quotePrefix        = []byte("> ")
	itemPrefix         = []byte("* ")
	itemIndent         = []byte{'\t'}
	preformattedToggle = []byte("```")
	codeDelimiter      = []byte("`")
	emphDelimiter      = []byte("*")
	strongDelimiter    = []byte("**")
	delDelimiter       = []byte("~~")
	horizontalRule     = []byte("---")
	subOpen            = []byte("_{")
	subClose           = []byte("}")
	supOpen            = []byte("^(")
	supClose           = []byte(")")
)

// matches a FULL string that contains no non-whitespace characters
var emptyLineRegex = regexp.MustCompile(`\A[\s]*\z`)

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

func getNodeDelimiter(node ast.Node) []byte {
	switch node.(type) {
	case *ast.Code:
		return codeDelimiter
	case *ast.Emph:
		return emphDelimiter
	case *ast.Strong:
		return strongDelimiter
	case *ast.Del:
		return delDelimiter
	default:
		return []byte{}
	}
}

func (r Renderer) link(w io.Writer, node *ast.Link, entering bool) {
	if entering {
		if node.Footnote != nil {
			fmt.Fprintf(w, "[^%d]: %s", node.NoteID, extractText(node.Footnote))
		} else {
			w.Write(linkPrefix)
			w.Write(node.Destination)
			w.Write(space)
			r.text(w, node)
		}
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

func (r Renderer) hr(w io.Writer, node *ast.HorizontalRule, entering bool) {
	if entering {
		w.Write(horizontalRule)
		w.Write(lineBreak)
		w.Write(lineBreak)
	}
}

// Based on https://pages.uoregon.edu/ncp/Courses/MathInPlainTextEmail.html
func (r Renderer) subscript(w io.Writer, node *ast.Subscript, entering bool) {
	if entering {
		if node := node.AsLeaf(); node != nil {
			w.Write(subOpen)
			w.Write([]byte(strings.ReplaceAll(string(node.Literal), "\n", " ")))
			w.Write(subClose)
		}
	}
}
func (r Renderer) superscript(w io.Writer, node *ast.Superscript, entering bool) {
	if entering {
		if node := node.AsLeaf(); node != nil {
			w.Write(supOpen)
			w.Write([]byte(strings.ReplaceAll(string(node.Literal), "\n", " ")))
			w.Write(supClose)
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

func extractLinks(node ast.Node) (stack []ast.Node) {
	if node := node.AsContainer(); node != nil {
		for _, subnode := range node.Children {
			stack = append(stack, extractLinks(subnode)...)
		}
	}
	switch node := node.(type) {
	case *ast.Image:
		stack = append(stack, node)
	case *ast.Link:
		stack = append(stack, node)
		// footnotes are represented as links which embed an extra node
		// containing footnote text; the link itself is not considered a
		// container
		if node.Footnote != nil {
			stack = append(stack, extractLinks(node.Footnote)...)
		}
	}
	return stack
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
				case *ast.Text, *ast.Code, *ast.Emph, *ast.Strong, *ast.Del, *ast.Link, *ast.Image:
					r.text(w, child)
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

func (r Renderer) code(w io.Writer, node *ast.CodeBlock) {
	w.Write(preformattedToggle)
	if node.IsFenced {
		w.Write(node.Info)
	}
	w.Write(lineBreak)
	w.Write(node.Literal)
	w.Write(preformattedToggle)
	w.Write(lineBreak)
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
			r.text(w, item)
			w.Write(lineBreak)
			if l >= 2 {
				if list, ok := item.Children[1].(*ast.List); ok {
					r.list(w, list, level+1)
				}
			}
		}
	}
}

var lineBreakCharacters = regexp.MustCompile(`[\n\r]+`)

func (r Renderer) text(w io.Writer, node ast.Node) {
	delimiter := getNodeDelimiter(node)
	// special case for footnotes: we want them in the text
	if node, ok := node.(*ast.Link); ok && node.Footnote != nil {
		fmt.Fprintf(w, "[^%d]", node.NoteID)
	}
	if node := node.AsLeaf(); node != nil {
		// replace all newlines in text with spaces, allowing for soft
		// wrapping; this is recommended as per Gemini spec p. 5.4.1
		w.Write(delimiter)
		w.Write(lineBreakCharacters.ReplaceAll(node.Literal, space))
		w.Write(delimiter)
		return
	}
	if node := node.AsContainer(); node != nil {
		w.Write(delimiter)
		for _, child := range node.Children {
			// skip non-text child elements from rendering
			switch child := child.(type) {
			case *ast.List:
			default:
				r.text(w, child)
			}
		}
		w.Write(delimiter)
	}
}

func (r Renderer) blockquoteText(w io.Writer, node ast.Node) {
	delimiter := getNodeDelimiter(node)
	if node, ok := node.(*ast.Link); ok && node.Footnote != nil {
		fmt.Fprintf(w, "[^%d]", node.NoteID)
	}
	if node := node.AsLeaf(); node != nil {
		// pad every line break with blockquote symbol
		w.Write(delimiter)
		w.Write(lineBreakCharacters.ReplaceAll(node.Literal, quoteBrPrefix))
		w.Write(delimiter)
		return
	}
	if node := node.AsContainer(); node != nil {
		w.Write(delimiter)
		for _, child := range node.Children {
			r.blockquoteText(w, child)
		}
		w.Write(delimiter)
	}
}

func extractText(node ast.Node) string {
	delimiter := getNodeDelimiter(node)
	if node := node.AsLeaf(); node != nil {
		return string(delimiter) + strings.ReplaceAll(string(node.Literal), "\n", " ") + string(delimiter)
	}
	if node := node.AsContainer(); node != nil {
		b := strings.Builder{}
		for _, child := range node.Children {
			b.WriteString(string(delimiter) + extractText(child) + string(delimiter))
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

// RenderNode implements Renderer.RenderNode().
func (r Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	// entering in gomarkdown was made to have elements of type switch
	// to enclose themselves within the second pass with entering =
	// false, as Markdown is quite similar to HTML in its structure.
	// As Gemtext is line-oriented, and not tag-oriented, most of
	// container subroutines have to handle their subelements on
	// themselves.
	noNewLine := true
	fetchLinks := false
	switch node := node.(type) {
	case *ast.BlockQuote:
		r.blockquote(w, node, entering)
		fetchLinks = true
	case *ast.HorizontalRule:
		r.hr(w, node, entering)
	case *ast.Heading:
		r.heading(w, node, entering)
		noNewLine = false
	case *ast.Paragraph:
		switch node.Parent.(type) {
		// these (should) handle underlying paragraphs themselves
		case *ast.BlockQuote, *ast.ListItem, *ast.Footnotes:
		default:
			noNewLine = r.paragraph(w, node, entering)
			fetchLinks = true
		}
	case *ast.CodeBlock:
		r.code(w, node)
		// code block is not considered a wrapping element
		w.Write(lineBreak)
	case *ast.List:
		// lists of level >= 2 are rendered recursively along with the
		// first level; the list is a container
		_, parentIsDocument := node.Parent.(*ast.Document)
		// footnotes are rendered as links after the parent paragraph
		if !node.IsFootnotesList && parentIsDocument && !entering {
			r.list(w, node, 0)
			noNewLine = false
			fetchLinks = true
		}
	case *ast.Table:
		r.table(w, node, entering)
		noNewLine = false
		fetchLinks = true
	}
	if !noNewLine && !entering {
		w.Write(lineBreak)
	}
	if fetchLinks && !entering {
		links := extractLinks(node)
		if len(links) > 0 {
			r.linksList(w, links)
		}
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
