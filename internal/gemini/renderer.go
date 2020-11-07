// Package gemini contains an implementation of markdown => text/gemini
// renderer for github.com/gomarkdown/markdown.
package gemini

import (
	"bufio"
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

var (
	lineBreak   = []byte{'\n'}
	space       = []byte{' '}
	linkPrefix  = []byte("=> ")
	quotePrefix = []byte("> ")
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

func (r Renderer) paragraph(w io.Writer, node *ast.Paragraph, entering bool) {
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
		w.Write(lineBreak)
		// render a links block after paragraph
		if len(linkStack) > 0 {
			w.Write(lineBreak)
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
}

func (r Renderer) code(w io.Writer, node *ast.Code, entering bool) {
	// TODO: Renderer.code: untested yet
	if entering {
		w.Write([]byte("```\n"))
		w.Write(node.Content)
	} else {
		w.Write([]byte("```\n"))
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
		if _, parentIsBlockQuote := node.Parent.(*ast.BlockQuote); !parentIsBlockQuote {
			r.paragraph(w, node, entering)
			noNewLine = false
		}
	case *ast.Code:
		// TODO: *ast.Code render is likely to be merged into paragraph
		r.code(w, node, entering)
		noNewLine = false
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
