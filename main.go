package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const TOC = "{{.TOC}}"

func main() {
	var (
		footerTemplateParm string
		footerTemplate     string
		markdownFile       string
	)

	pflag.StringVar(&markdownFile, "markdown", "", "markdown file")
	pflag.StringVar(&footerTemplateParm, "footer-template", "", "Footer template")

	// Parse flags
	pflag.Parse()

	if !fileExists(markdownFile) {
		log.Printf("File does not exist: %s\n", markdownFile)
		os.Exit(1)
	}

	// Conditionally add FooterTemplate
	if (footerTemplateParm != "") && !fileExists(footerTemplateParm) {
		log.Printf("Footer template not found: %v\n", footerTemplateParm)
		footerTemplateParm = ""
	}
	if (footerTemplateParm != "") && fileExists(footerTemplateParm) {
		content, err := os.ReadFile(footerTemplateParm)
		if err != nil {
			log.Printf("Failed to read footer template file: %v", err)
			footerTemplateParm = ""
			footerTemplate = ""
		} else {
			footerTemplate = string(content)
			log.Printf("Using Footer template: %v\n", footerTemplateParm)
		}
	}

	markdownContent, err := os.ReadFile(markdownFile)
	if err != nil {
		log.Printf("Failed to read markdown file: %v", err)
		os.Exit(2)
	}
	markdownContent = insertTOC(markdownContent)

	// Convert markdown to HTML (you can use any markdown to HTML converter)
	htmlContent := convertMarkdownToHTML(string(markdownContent))
	tempHTMLFile, err := os.Create("md-html-temp.html")
	if err != nil {
		log.Printf("Failed to create temporary HTML file: %v", err)
		os.Exit(3)
	}
	defer os.Remove(tempHTMLFile.Name()) // Clean up the temporary file

	// Write the HTML content to the temporary file
	_, err = tempHTMLFile.Write([]byte(htmlContent))
	if err != nil {
		log.Printf("Failed to write to temporary HTML file: %v", err)
		os.Exit(5)
	}
	tempHTMLFile.Close()

	// Get the full path of the temporary HTML file
	fullPath, err := filepath.Abs(tempHTMLFile.Name())
	if err != nil {
		log.Printf("Failed to get absolute path: %v", err)
		os.Exit(6)
	}

	// Create a context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Run chromedp tasks
	var pdfBuffer []byte
	uri := "file://" + fullPath
	err = chromedp.Run(ctx, generatePDF(uri, &pdfBuffer, footerTemplate))
	if err != nil {
		fmt.Printf("Failed to generate file: %v \n", uri)
		log.Fatalf("Failed to generate PDF: %v\n", err)
	}

	// Write the PDF to a file
	// Strip extension from markdown file

	pdfFile := stripExtension(markdownFile) + ".pdf"
	err = os.WriteFile(pdfFile, pdfBuffer, 0644)
	if err != nil {
		log.Fatalf("Failed to write PDF file: %v - ", err)
		os.Exit(7)
	}
	// delete tempHTMLFile
	err = os.Remove(tempHTMLFile.Name())
	if err != nil {
		log.Printf("Failed to delete tempHTMLFile: %v - ", err)
		os.Exit(8)
	}

	fmt.Printf("PDF file generated: %s\n", pdfFile)
}

func convertMarkdownToHTML(markdownInput string) []byte {
	md := []byte(markdownInput)
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

// generatePDF generates a PDF from the given URL and stores it in the res variable
// see https://github.com/chromedp/cdproto/blob/master/page/page.go
func generatePDF(urlstr string, res *[]byte, footertemplate string) chromedp.Tasks {
	marginTop := float64(0.8)    // 0.4 inch margin,
	marginBottom := float64(0.8) // 0.4 inch margin,
	marginLeft := float64(0.8)   // 0.4 inch margin,
	marginRight := float64(0.8)  // 0.4 inch margin,
	// A4 paper size
	paperWidth := float64(8.27)   // 8.5 inch
	paperHeight := float64(11.69) // 11 inch

	printToPDF := page.PrintToPDF().
		WithPrintBackground(false).
		WithMarginTop(marginTop).
		WithMarginBottom(marginBottom).
		WithMarginLeft(marginLeft).
		WithMarginRight(marginRight).
		WithPaperWidth(paperWidth).
		WithPaperHeight(paperHeight)

	// Conditionally add FooterTemplate
	if footertemplate != "" {
		printToPDF = printToPDF.WithFooterTemplate(footertemplate).WithDisplayHeaderFooter(true)
		printToPDF = printToPDF.WithHeaderTemplate("<span class=title></span>")

	}

	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := printToPDF.Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func stripExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

func insertTOC(markdownContent []byte) []byte {
	// Parse the markdown content
	parser := parser.New()
	doc := markdown.Parse(markdownContent, parser)

	// Generate the TOC
	toc := generateTOC(doc)

	// replace TOC placeholder with the generated TOC
	markdownContent = bytes.Replace(markdownContent, []byte(TOC), []byte(toc), 1)
	return markdownContent
}

func generateTOC(doc ast.Node) string {
	var tocBuilder strings.Builder
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if header, ok := node.(*ast.Heading); ok {
			if header.Level > 3 {
				return ast.GoToNext
			}

			text := extractText(header)
			anchor := generateAnchor(text)
			tocBuilder.WriteString(fmt.Sprintf("%s- [%s](#%s)\n", strings.Repeat("  ", header.Level-1), text, anchor))
		}

		return ast.GoToNext
	})

	tocBuilder.WriteString("\n")
	return tocBuilder.String()
}

func extractText(header *ast.Heading) string {
	var textBuilder strings.Builder
	for _, node := range header.Children {
		if text, ok := node.(*ast.Text); ok {
			textBuilder.Write(text.Literal)
		}
	}
	return textBuilder.String()
}

func generateAnchor(text string) string {
	return strings.ToLower(strings.ReplaceAll(text, " ", "-"))
}
