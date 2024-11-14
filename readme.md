# Markdown to pdf

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
