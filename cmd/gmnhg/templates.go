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
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/tdemin/gmnhg/internal/gmnhg"
)

func mustParseTmpl(name, value string) *template.Template {
	return template.Must(template.New(name).Funcs(defineFuncMap()).Parse(value))
}

func defineFuncMap() template.FuncMap {
	fm := sprig.TxtFuncMap()
	// sorts posts by date, newest posts go first
	fm["sortPosts"] = gmnhg.SortRev
	fm["sort"] = gmnhg.Sort
	fm["sortRev"] = gmnhg.SortRev
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
