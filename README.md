# Hugo-to-Gemini converter

This repo holds a converter of Hugo Markdown posts to
[text/gemini][Gemtext] (also named Gemtext in this README). The
converter is supposed to make people using [Hugo](https://gohugo.io)'s
entrance to [Project Gemini][Gemini], the alternate web, somewhat
simpler.

[Gemini]: https://gemini.circumlunar.space
[Gemtext]: https://gemini.circumlunar.space/docs/specification.html

At this stage of development this repo contains the actual renderer
(`internal/gemini`) and the `md2gmn` program that converts Markdown
input to Gemtext and is supposed to facilitate testing.

The renderer is somewhat hasty, and is NOT supposed to be able to
convert the entirety of possible Markdown to Gemtext (as it's not
possible to do so, considering Gemtext is a lot simpler than Markdown),
but instead a selected subset of it, enough for conveying your mind in
Markdown.

The renderer uses the [gomarkdown][gomarkdown] library for parsing
Markdown.

[gomarkdown]: https://github.com/gomarkdown/markdown

## md2gmn

This program reads Markdown input from either text file (if `-f
filename` is given), or stdin. The resulting Gemtext goes to stdout.

```
Usage of md2gmn:
  -f string
        input file
```

## TODO

+ [x] convert Markdown text to Gemtext
+ [ ] prepend contents of YAML front matter to Gemtext data
+ [ ] render all Hugo content files to Gemtext in accordance with front
  matter data and Hugo config

## License

This program is redistributed under the terms and conditions of the GNU
General Public License, more specifically under version 3 of the
License. For details, see [COPYING](COPYING).
