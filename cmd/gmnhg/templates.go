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

var defaultSingleTemplate = mustParseTmpl("single", `# {{ .Metadata.Title }}

{{ .Metadata.Date.Format "2006-01-02 15:04" }}

{{ printf "%s" .Post }}`)

var defaultIndexTemplate = mustParseTmpl("index", `# Site index

{{ with .Content }}{{ printf "%s" . -}}{{ end }}
{{- range $dir, $posts := .PostData }}{{ if and (ne $dir "/") (eq (dir $dir) "/") }}
Index of {{ trimPrefix "/" $dir }}:

{{ range $p := $posts | sortPosts }}=> {{ $p.Link }} {{ $p.Metadata.Date.Format "2006-01-02 15:04" }} - {{ if $p.Metadata.Title }}{{ $p.Metadata.Title }}{{else}}{{ $p.Link }}{{end}}
{{ end }}{{ end }}{{ end }}
`)

var defaultRssTemplate = mustParseTmpl("rss", `{{- $Site := .Site -}}
{{- $SiteTitle := or $Site.GmnhgTitle $Site.Title | html -}}
{{- $SiteBaseURL := or $Site.GmnhgBaseURL $Site.BaseURL | trimSuffix "/" | html -}}
{{- $Dirname := .Dirname | trimPrefix "/" | html -}}
{{- $DirURL := list $SiteBaseURL $Dirname | join "/" | html -}}
{{- $RssURL := list $SiteBaseURL (trimPrefix "/" .Link) | join "/" | html -}}
{{- $RssTitle := printf "%s%s" (or $SiteTitle "Site feed") (and $Dirname (printf " - %s" $Dirname)) | html -}}
<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>{{ $RssTitle }}</title>
    <link>{{ $DirURL }}</link>
    <description>Recent content{{ with $Dirname }} in {{ . }}{{end}}{{ with $SiteTitle }} on {{ . }}{{end}}</description>
    <generator>gmnhg</generator>{{ with $Site.LanguageCode }}
    <language>{{ html .}}</language>{{end}}{{ with $Site.Author.email }}
    <managingEditor>{{ html . }}{{ with $Site.Author.name }} ({{ html . }}){{end}}</managingEditor>
    <webMaster>{{ html . }}{{ with $Site.Author.name }} ({{ html . }}){{end}}</webMaster>{{end}}{{ with $Site.Copyright }}
    <copyright>{{ html . }}</copyright>{{end}}
    <lastBuildDate>{{ now.Format "Mon, 02 Jan 2006 15:04:05 -0700" }}</lastBuildDate>
    {{ printf "<atom:link href=%q rel=\"self\" type=\"application/rss+xml\" />" $RssURL }}
    {{ range $i, $p := .Posts | sortPosts }}{{ if lt $i 25 }}
    {{- $RelURL := trimPrefix "/" $p.Link | html -}}
    {{- $AbsURL := list $SiteBaseURL $RelURL | join "/" }}
    <item>
      <title>{{ if $p.Metadata.Title }}{{ html $p.Metadata.Title }}{{ else }}{{ $RelURL }}{{end}}</title>
      <link>{{ $AbsURL }}</link>
      <pubDate>{{ $p.Metadata.Date.Format "Mon, 02 Jan 2006 15:04:05 -0700" }}</pubDate>
      <guid>{{ $AbsURL }}</guid>
      <description>{{ html $p.Metadata.Summary }}</description>
    </item>
    {{end}}{{end}}
  </channel>
</rss>
`)
