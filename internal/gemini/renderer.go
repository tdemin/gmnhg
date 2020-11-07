// Package gemini contains an implementation of markdown => text/gemini
// renderer for github.com/gomarkdown/markdown.
package gemini

import (
	"io"

	"github.com/gomarkdown/markdown/ast"
)

var (
	lineBreak          = []byte{'\n'}
	lineBreakByte byte = 0x0a
	space              = []byte{' '}
	linkPrefix         = []byte("=> ")
	quotePrefix        = []byte("> ")
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
		if node.Title != nil {
			w.Write(space)
			w.Write(node.Title)
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
		w.Write(quotePrefix)
		if para, ok := node.Children[0].(*ast.Paragraph); ok {
			for _, subnode := range para.Children {
				if l := subnode.AsLeaf(); l != nil {
					// TODO: Renderer.blockquote: rendering by byte is asking
					// for optimizations
					for _, b := range l.Literal {
						w.Write([]byte{b})
						if b == lineBreakByte {
							w.Write(quotePrefix)
						}
					}
				}
			}
		}
	} else {
		w.Write(lineBreak)
		w.Write(lineBreak)
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
		w.Write(lineBreak)
	}
}

func (r Renderer) paragraph(w io.Writer, node *ast.Paragraph, entering bool) {
	if entering {
		linkStack := make([]ast.Node, 0, len(node.Children))
		onlyElement := len(node.Children) == 1
		for _, child := range node.Children {
			// only render links text in the paragraph if they're
			// combined with some other text on page
			if link, ok := child.(*ast.Link); ok {
				if !(len(link.Children) > 1) && !onlyElement {
					r.linkText(w, link)
				}
				linkStack = append(linkStack, link)
			}
			if image, ok := child.(*ast.Image); ok {
				if !(len(image.Children) > 1) && !onlyElement {
					r.imageText(w, image)
				}
				linkStack = append(linkStack, image)
			}
			if text, ok := child.(*ast.Text); ok {
				r.text(w, text)
			}
		}
		// render a links block after paragraph
		if len(linkStack) > 0 {
			w.Write(lineBreak)
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
	} else {
		w.Write(lineBreak)
		w.Write(lineBreak)
	}
}

func (r Renderer) code(w io.Writer, node *ast.Code, entering bool) {
	// TODO: Renderer.code: no good way to test that yet
	if entering {
		w.Write([]byte("```\n"))
		w.Write(node.Content)
	} else {
		w.Write([]byte("```\n\n"))
	}
}

func (r Renderer) text(w io.Writer, node *ast.Text) {
	w.Write(node.Literal)
}

// RenderNode implements Renderer.RenderNode().
func (r Renderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	// despite most of the subroutines here accepting entering, most of
	// them don't really need an extra pass
	switch node := node.(type) {
	case *ast.BlockQuote:
		r.blockquote(w, node, entering)
	case *ast.Heading:
		r.heading(w, node, entering)
	case *ast.Paragraph:
		// blockquote wraps paragraphs which makes for an extra render
		if _, parentIsBlockQuote := node.Parent.(*ast.BlockQuote); !parentIsBlockQuote {
			r.paragraph(w, node, entering)
		}
	case *ast.Code:
		// TODO: *ast.Code render is likely to be merged into paragraph
		r.code(w, node, entering)
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
