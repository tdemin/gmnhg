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

// Package gemini provides functions to convert Markdown files to
// Gemtext. It supports the use of YAML front matter in Markdown.
package gemini

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/tdemin/gmnhg/internal/gemini"
	"gopkg.in/yaml.v2"
)

// HugoMetadata implements gemini.Metadata, providing the bare minimum
// of possible post props.
type HugoMetadata struct {
	PostTitle   string    `yaml:"title"`
	PostIsDraft bool      `yaml:"draft"`
	PostLayout  string    `yaml:"layout"`
	PostDate    time.Time `yaml:"date"`
}

// Title returns post title.
func (h HugoMetadata) Title() string {
	return h.PostTitle
}

// Date returns post date.
func (h HugoMetadata) Date() time.Time {
	return h.PostDate
}

var (
	yamlDelimiter = []byte("---\n")
	trailing      = []byte("\n\n")
)

// ErrPostIsDraft indicates the post rendered is a draft and is not
// supposed to be rendered.
var ErrPostIsDraft = errors.New("post is draft")

// Settings is a bitmask for renderer preferences.
type Settings uint

// Has uses AND to check whether a flag is set.
func (s Settings) Has(setting Settings) bool {
	return (s & setting) != 0
}

const (
	// Defaults simply renders the document.
	Defaults Settings = 0
	// WithMetadata indicates that the metadata should be included in
	// the text produced by the renderer.
	WithMetadata Settings = 1 << iota
)

// RenderMarkdown converts Markdown text to Gemtext using gomarkdown. It
// appends Hugo YAML front matter data to the post header if
// WithMetadata is set.
//
// Only a subset of front matter data parsed by Hugo is included in the
// final document. At this point it's just title and date.
//
// Draft posts are still rendered, but with an error of type
// ErrPostIsDraft.
func RenderMarkdown(md []byte, settings Settings) (geminiText []byte, metadata HugoMetadata, err error) {
	var (
		blockEnd    int
		yamlContent []byte
	)
	// only allow front matter at file start
	if bytes.Index(md, yamlDelimiter) != 0 {
		goto parse
	}
	blockEnd = bytes.Index(md[len(yamlDelimiter):], yamlDelimiter)
	if blockEnd == -1 {
		goto parse
	}
	yamlContent = md[len(yamlDelimiter) : blockEnd+len(yamlDelimiter)]
	if err := yaml.Unmarshal(yamlContent, &metadata); err != nil {
		return nil, metadata, fmt.Errorf("invalid front matter: %w", err)
	}
	md = md[blockEnd+len(yamlDelimiter)*2:]
parse:
	ast := markdown.Parse(md, parser.NewWithExtensions(parser.CommonExtensions|
		parser.NoEmptyLineBeforeBlock|
		parser.Footnotes))
	var content []byte
	if settings.Has(WithMetadata) && metadata.PostTitle != "" {
		content = markdown.Render(ast, gemini.NewRendererWithMetadata(metadata))
	} else {
		content = markdown.Render(ast, gemini.NewRenderer())
	}
	for li := bytes.LastIndex(content, trailing); li != -1; li = bytes.LastIndex(content, trailing) {
		if li != len(content)-len(trailing) {
			break
		}
		content = content[:len(content)-1]
	}
	if metadata.PostIsDraft {
		return content, metadata, fmt.Errorf("%s: %w", metadata.PostTitle, ErrPostIsDraft)
	}
	return content, metadata, nil
}
