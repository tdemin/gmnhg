// Package gemini contains an implementation of markdown => text/gemini
// renderer for github.com/gomarkdown/markdown.
package gemini

import (
	"io"

	"github.com/gomarkdown/markdown/ast"
)

var (
	lineBreak  = []byte{'\n'}
	space      = []byte{' '}
	linkPrefix = []byte("=> ")
)

// Renderer implements markdown.Renderer.
type Renderer struct {
	LinkStack []ast.Node
}

// NewRenderer returns a new Renderer.
func NewRenderer() Renderer {
	return Renderer{
		LinkStack: nil,
	}
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

func (r Renderer) citation(w io.Writer, node *ast.Citation) {}

func (r Renderer) heading(w io.Writer, node *ast.Heading, entering bool) {
	if entering {
		// prepend headings with the relevant number of #-s
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
		if r.LinkStack != nil {
			panic("link stack not empty")
		}
		r.LinkStack = make([]ast.Node, 0, len(node.Children))
		for _, child := range node.Children {
			if link, ok := child.(*ast.Link); ok {
				r.LinkStack = append(r.LinkStack, link)
			}
			if image, ok := child.(*ast.Image); ok {
				r.LinkStack = append(r.LinkStack, image)
			}
			if text, ok := child.(*ast.Text); ok {
				r.text(w, text)
			}
		}
	} else {
		w.Write(lineBreak)
		w.Write(lineBreak)
		for _, link := range r.LinkStack {
			if link, ok := link.(*ast.Link); ok {
				r.link(w, link, true)
			}
			if image, ok := link.(*ast.Image); ok {
				r.image(w, image, true)
			}
			w.Write(lineBreak)
		}
		r.LinkStack = nil
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
	case *ast.Link:
		// TODO: shouldn't be here at all
		r.link(w, node, entering)
	case *ast.Image:
		// TODO: neither this
		r.image(w, node, entering)
	case *ast.Citation:
		// TODO: neither this
		r.citation(w, node)
	case *ast.Heading:
		r.heading(w, node, entering)
	case *ast.Paragraph:
		r.paragraph(w, node, entering)
	case *ast.Code:
		// TODO: likely not even this
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
