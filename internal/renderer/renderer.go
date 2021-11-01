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
	"fmt"
	"io"
	"regexp"

	"github.com/gomarkdown/markdown/ast"
)

var (
	lineBreak       = []byte{'\n'}
	space           = []byte{' '}
	linkPrefix      = []byte("=> ")
	codeDelimiter   = []byte("`")
	emphDelimiter   = []byte("*")
	strongDelimiter = []byte("**")
	delDelimiter    = []byte("~~")
)

// matches a FULL string that contains no non-whitespace characters
var emptyLineRegex = regexp.MustCompile(`\A[\s]*\z`)

var lineBreakCharacters = regexp.MustCompile(`[\n\r]+`)

// Renderer implements markdown.Renderer.
type Renderer struct{}

// NewRenderer returns a new Renderer.
func NewRenderer() Renderer {
	return Renderer{}
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

func textWithNewlineReplacement(node ast.Node, replacement []byte, unescapeHtml bool) []byte {
	buf := bytes.Buffer{}
	delimiter := getNodeDelimiter(node)
	// special case for footnotes: we want them in the text
	if node, ok := node.(*ast.Link); ok && node.Footnote != nil {
		fmt.Fprintf(&buf, "[^%d]", node.NoteID)
	}
	if leaf := node.AsLeaf(); leaf != nil {
		// replace all newlines in text with preferred symbols; this may
		// be spaces for general text, allowing for soft wrapping, which
		// is recommended as per Gemini spec p. 5.4.1, or line breaks
		// with a blockquote symbols for blockquotes, or just nothing
		buf.Write(delimiter)
		switch node := node.(type) {
		case *ast.Hardbreak:
			buf.Write(lineBreak)
			// If the blockquote ends with a double space, the parser will
			// not create a Hardbreak at the end, so this works.
			if _, ok := leaf.Parent.(*ast.BlockQuote); !ok {
				buf.Write(quotePrefix)
			}
		case *ast.HTMLSpan:
			if isHardBreak(leaf.Literal) {
				buf.Write(lineBreak)
			}
			buf.Write(leaf.Content)
		case *ast.HTMLBlock:
			buf.Write([]byte(stripHtml(node, quotePrefix)))
		default:
			textWithoutBreaks := lineBreakCharacters.ReplaceAll(leaf.Literal, replacement)
			if unescapeHtml {
				buf.Write(unescapeHtmlText(textWithoutBreaks))
			} else {
				buf.Write(textWithoutBreaks)
			}
		}
		buf.Write(delimiter)
	}
	if node := node.AsContainer(); node != nil {
		buf.Write(delimiter)
		for _, child := range node.Children {
			// skip non-text child elements from rendering
			switch child := child.(type) {
			case *ast.List:
			default:
				buf.Write(textWithNewlineReplacement(child, replacement, unescapeHtml))
			}
		}
		buf.Write(delimiter)
	}
	return buf.Bytes()
}

func (r Renderer) text(w io.Writer, node ast.Node, unescapeHtml bool) {
	w.Write(textWithNewlineReplacement(node, space, unescapeHtml))
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

func isLinksOnlyList(node *ast.List) bool {
	for _, child := range node.Children {
		child, ok := child.(*ast.ListItem)
		if !ok {
			return false // should never happen
		}
		for _, liChild := range child.Children {
			liChild, ok := liChild.(*ast.Paragraph)
			if !ok {
				return false // sublist, etc
			}
			if !isLinksOnlyParagraph(liChild) {
				return false
			}
		}
	}
	return true
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
			if !isLinksOnlyList(node) {
				r.list(w, node, 0)
				noNewLine = false
			}
			fetchLinks = true
		}
	case *ast.Table:
		r.table(w, node, entering)
		noNewLine = false
		fetchLinks = true
	case *ast.HTMLBlock:
		// Do not render if already rendered as part of a blockquote
		if _, ok := node.Parent.(*ast.BlockQuote); !ok {
			r.htmlBlock(w, node, entering)
		}
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

// RenderHeader implements Renderer.RenderHeader().
func (r Renderer) RenderHeader(w io.Writer, node ast.Node) {}

// RenderFooter implements Renderer.RenderFooter().
func (r Renderer) RenderFooter(w io.Writer, node ast.Node) {}
