// Typst settings for Manual.pdf (pandoc --include-in-header).

#let accent = rgb("#1f4e79")   // headings
#let accent2 = rgb("#2e75b6")  // rules, code block edge
#let codebg = rgb("#f3f6fa")   // code block background
#let typed = rgb("#0b5394")    // input lines in examples
#let caret = rgb("#c00000")    // error caret
#let inline = rgb("#9c3d10")   // inline code
#let headfill = rgb("#dce6f2") // table header
#let stripe = rgb("#f5f8fc")   // table rows

// Long tables may break across pages.
#show figure: set block(breakable: true)
#show table: set par(justify: false)
#set table(
  stroke: none,
  fill: (x, y) => if y == 0 { headfill } else if calc.even(y) { stripe },
)

#show heading: set text(fill: accent)
#show heading.where(level: 1): it => block(
  width: 100%,
  above: 1.6em,
  below: 0.9em,
  stroke: (bottom: 0.8pt + accent2),
  inset: (bottom: 4pt),
  it.body,
)

#show raw: set text(font: "DejaVu Sans Mono")
#show raw.where(block: false): set text(fill: inline)
#show raw.where(block: true): it => block(
  width: 100%,
  fill: codebg,
  inset: (x: 8pt, y: 6pt),
  radius: 2pt,
  stroke: (left: 2pt + accent2),
  it,
)

// Session examples: typed lines in color, error carets in red.
#show raw.where(block: true, lang: "apl"): it => {
  show raw.line: l => {
    if l.text.starts-with("      ") or l.text.match(regex("^\[[\d.]+\] ([^ ]|$)")) != none {
      text(fill: typed, l.body)
    } else if l.text.trim() == "^" {
      text(fill: caret, l.body)
    } else {
      l.body
    }
  }
  it
}
