# Tables

gmnhg uses preformatted text blocks to render ASCII text tables.

## Simple table example

| Syntax      | Description |
| ----------- | ----------- |
| Header      | Title       |
| Paragraph   | Text        |

## Empty rows or cells

These are picked up as well.

| test | nice |
|------|------|
| `est` | |

| test | nice |
|------|------|
| | |

## Formatting inside tables

Text formatting is fully supported inside tables. Links will also get
picked up, and a links block will appear after the parent table if
needed.

| Header 1 | [Header 2](https://example.tld) | Header 3[^foo] |
|----------|----------|----------|
| Item 1   | [Item 2](https://www.example.com)   | Item 3   |
| Item 1a  | Item 2a  | Item 3a  |

[^foo]: Example footnote that explains header 3.
