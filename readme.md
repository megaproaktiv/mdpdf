# Markdown to pdf

What problem does this solve?

I often use the previewer enhanced from VSCode to create PDF Files from markdown files.
But wen i want to automate the creation of several markdown file pdfs at once, it gets tricky.

Using *pandoc* you need to install latex. It will look real nice, but it takes time.
I tried some other solutions but non fit.

So the idea was to use the Chrome Chrome DevTools Protocol. This fast small go app calls chrome to render markdown to pdf, so that you can automate it from the command line. 

## Usage

You need to have chrome installed.

Call app with the markdown file as parameter.

```bash
mdpdf --markdown changelog.md
```

## Footer

Parameter `-footer-template` can be used to define a footer template. The template can contain the following placeholders:

  date: formatted print date
  title: document title
  url: document location
  pageNumber: current page number
  totalPages: total pages in the document

Example for a footer file:

```html
<div class="page-footer" style="width:100%; text-align:right; padding-right:10px; font-size:12px;"><span class="pageNumber"></span> / <span class="totalPages"></span></div>
```
