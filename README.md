# Hugo-to-Gemini converter

[![PkgGoDev](https://pkg.go.dev/badge/github.com/tdemin/gmnhg)](https://pkg.go.dev/github.com/tdemin/gmnhg)

This repo holds a converter of Hugo Markdown posts to
[text/gemini][Gemtext] (also named Gemtext in this README). The
converter is supposed to make people using [Hugo](https://gohugo.io)'s
entrance to [Project Gemini][Gemini], the alternate web, somewhat
simpler.

[Gemini]: https://gemini.circumlunar.space
[Gemtext]: https://gemini.circumlunar.space/docs/specification.html

The renderer is somewhat hasty, and is NOT supposed to be able to
convert the entirety of possible Markdown to Gemtext (as it's not
possible to do so, considering Gemtext is a lot simpler than Markdown),
but instead a selected subset of it, enough for conveying your mind in
Markdown.

The renderer uses the [gomarkdown][gomarkdown] library for parsing
Markdown. gomarkdown has a few quirks at this time, the most notable one
being unable to parse links/images inside other links.

[gomarkdown]: https://github.com/gomarkdown/markdown

## gmnhg

This program converts Hugo Markdown content files from `content/` in
accordance with templates found in `gmnhg/` to the output dir. It
also copies static files from `static/` to the output dir.

For more details about the rendering process, see the
[doc](cmd/gmnhg/main.go) attached to the program.

```
Usage of gmnhg:
  -output string
        output directory (will be created if missing) (default "output/")
  -working string
        working directory (defaults to current directory)
```

## md2gmn

This program reads Markdown input from either text file (if `-f
filename` is given), or stdin. The resulting Gemtext goes to stdout.

```
Usage of md2gmn:
  -f string
        input file
```

md2gmn is mainly made to facilitate testing the Gemtext renderer but
can be used as a standalone program as well.

## License

This program is redistributed under the terms and conditions of the GNU
General Public License, more specifically version 3 of the License. For
details, see [COPYING](COPYING).
