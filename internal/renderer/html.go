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
	"html"
	"io"
	"regexp"

	"github.com/gomarkdown/markdown/ast"
	"github.com/grokify/html-strip-tags-go"
)

// fairly tolerant to handle weird HTML
var tagPairRegexString = `<[\n\f ]*%s([\n\f ]+[^\n\f \/>"'=]+[\n\f ]*(=[\n\f ]*([a-zA-Z1-9\-]+|"[^\n\f"]+"|'[^\n\f']+'))?)*[\n\f ]*>.*?<[\n\f ]*/[\n\f ]*%s[\n\f ]*>`

// HTML block tags whose contents should not be rendered
var htmlNoRenderRegex = []*regexp.Regexp{
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "fieldset", "fieldset")),
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "form", "form")),
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "iframe", "iframe")),
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "script", "script")),
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "style", "style")),
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "canvas", "canvas")),
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "dialog", "dialog")),
	regexp.MustCompile(fmt.Sprintf(tagPairRegexString, "progress", "progress")),
}

var hardBreakTag = regexp.MustCompile(`< *br */? *>`)
var escapedHtmlChar = regexp.MustCompile(`(?:^|[^\\\\])&[[:alnum:]]+;`)

func (r Renderer) htmlBlock(w io.Writer, node *ast.HTMLBlock, entering bool) {
	if entering {
		htmlString := stripHtml(node, []byte{})
		if len(htmlString) > 0 {
			w.Write([]byte(htmlString))
			w.Write(lineBreak)
			w.Write(lineBreak)
		}
	}
}

func stripHtml(node *ast.HTMLBlock, linePrefix []byte) string {
	// Only render contents of allowed tags
	literal := node.Literal
	for _, re := range htmlNoRenderRegex {
		literal = re.ReplaceAllLiteral(literal, []byte{})
	}
	if len(literal) > 0 {
		literalWithBreaks := hardBreakTag.ReplaceAll(lineBreakCharacters.ReplaceAll(literal, space), append([]byte(lineBreak), linePrefix...))
		literalStripped := strip.StripTags(string(literalWithBreaks))
		return html.UnescapeString(literalStripped)
	}
	return ""
}

func unescapeHtmlText(text []byte) []byte {
	return escapedHtmlChar.ReplaceAll(text, []byte(html.UnescapeString(string(text))))
}

func isHardBreak(text []byte) bool {
	return hardBreakTag.Match(text)
}
