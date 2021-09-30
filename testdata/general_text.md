# General text

Paragraphs are printed verbatim in gmnhg.

Single newlines (like in this multi-line paragraph) will get replaced by
a space, as Gemini specification p. 5.4.1 recommends this for
soft-wrapping text by clients.

Inline formatting bits (like this **bold** text, _emphasized_ text,
~~strikethrough~~ text, `preformatted text`) are kept to make sure
Gemini readers still have the stylistic context of your text.

## Blockquotes

Newlines in blockquote paragraphs, unlike usual paragraphs, aren't
replaced with a space. This facilitates appending authorship information
to the quote, or using blockquotes to write poems.

> "Never trouble another for what you can do yourself"
> — Thomas Jefferson, 3rd president of the US

> "Wow, writing comprehensive test suites is hard!"
> — Timur Demin, while writing this very test file

> "Somehow I know these two paragraphs will be broken into two separate
> blockquotes by gmnhg. I think my knowledge of that comes from being
> the author of this program."
>
> — also Timur Demin, in the process of writing this test file

## Code

gmnhg will use Gemtext preformatted blocks for that. Markdown alt-text
for preformatted blocks is supported, and is used to render alt-text as
specified by Gemini spec p. 5.4.3.

```go
package main

func main() {
    println("gmnhg is awesome!")
}
```

Preformatted Markdown of course isn't rendered:

```
# I am a test Markdown document

I contain text in **bold**.
```

## Links

gmnhg supports links, images, and footnotes. Links are a very
interesting topic on itself; see a [separate document](links.md) for
those.

## Lists

Definition lists, numbered and ordered lists are all supported in gmnhg.
There's also a [separate document](lists.md) displaying those.

## Tables

Markdown tables are supported in gmnhg, and are better displayed by a
[separate document](tables.md).

## Headings

Gemini specification allows up to three heading levels, with an optional
space after the last heading symbol, `#`. With Markdown, you get 6;
gmnhg will simply print the relevant number of #-s, making the client up
to parse more heading levels and keeping context of the source document.

Since clients like Lagrange treat the fourth and the rest of #-s as
heading content, it's best to avoid using H4-H6 in Gemini-aware Markdown
entirely. Headings from H3 to H6 are provided below so you can test how
your client handles that.

### Heading 3

#### Heading 4

##### Heading 5

###### Heading 6

## Misc

Inline HTML is <span class="bold">currently</span> stripped, but HTML
contents remain on-screen. This may change in the future.

> There's currently a [bug in gmnhg][bug] which prevents it from
> stripping HTML in certain scenarios. HTML is noticeably still present
> inside <span>blockquotes</span>.

***

The Markdown horizontal line above is rendered as triple dashes.

[bug]: https://github.com/tdemin/gmnhg/issues/6