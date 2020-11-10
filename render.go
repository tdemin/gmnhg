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
	"fmt"
	"time"

	"git.tdem.in/tdemin/gmnhg/internal/gemini"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v2"
)

// hugoMetadata implements gemini.Metadata, providing the bare minimum
// of possible post props.
type hugoMetadata struct {
	PostTitle string    `yaml:"title"`
	PostDate  time.Time `yaml:"date"`
}

func (h hugoMetadata) Title() string {
	return h.PostTitle
}

func (h hugoMetadata) Date() time.Time {
	return h.PostDate
}

var yamlDelimiter = []byte("---\n")

// RenderMarkdown converts Markdown text to text/gemini using
// gomarkdown, appending Hugo YAML front matter data if any is present
// to the post header.
//
// Only a subset of front matter data parsed by Hugo is included in the
// final document. At this point it's just title and date.
func RenderMarkdown(md []byte) (geminiText []byte, err error) {
	var (
		metadata    hugoMetadata
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
		return nil, fmt.Errorf("invalid front matter: %w", err)
	}
	md = md[blockEnd+len(yamlDelimiter)*2:]
parse:
	ast := markdown.Parse(md, parser.NewWithExtensions(parser.CommonExtensions))
	var geminiContent []byte
	if metadata.PostTitle != "" {
		geminiContent = markdown.Render(ast, gemini.NewRendererWithMetadata(metadata))
	} else {
		geminiContent = markdown.Render(ast, gemini.NewRenderer())
	}
	return geminiContent, nil
}
