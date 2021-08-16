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

package main

import (
	"github.com/Masterminds/sprig/v3"
	"sort"
	"text/template"
)

type postsSort []*post

func (p postsSort) Len() int {
	return len(p)
}

func (p postsSort) Less(i, j int) bool {
	return p[i].Metadata.PostDate.After(p[j].Metadata.PostDate)
}

func (p postsSort) Swap(i, j int) {
	t := p[i]
	p[i] = p[j]
	p[j] = t
}

func mustParseTmpl(name, value string) *template.Template {
	return template.Must(template.New(name).Funcs(defineFuncMap()).Parse(value))
}

func defineFuncMap() template.FuncMap {
	fm := sprig.TxtFuncMap()	
	// sorts posts by date, newest posts go first
	fm["sortPosts"] = func(posts []*post) []*post {
		// sortPosts is most likely to be used in a pipeline, and the
		// user has every right to expect it doesn't modify their
		// existing posts slice
		ps := make(postsSort, len(posts))
		copy(ps, posts)
		sort.Sort(ps)
		return ps
	}
	return fm
}

var defaultSingleTemplate = mustParseTmpl("single", `# {{ .Metadata.PostTitle }}

{{ .Metadata.PostDate.Format "2006-01-02 15:04" }}

{{ printf "%s" .Post }}`)

var defaultIndexTemplate = mustParseTmpl("index", `# Site index

{{ with .Content }}{{ printf "%s" . -}}{{ end }}
{{ range $dir, $posts := .PostData }}Index of {{ $dir }}:

{{ range $p := $posts | sortPosts }}=> {{ $p.Link }} {{ $p.Metadata.PostDate.Format "2006-01-02 15:04" }} - {{ $p.Metadata.PostTitle }}
{{ end }}{{ end }}
`)
