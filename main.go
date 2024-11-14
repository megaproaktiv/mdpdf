package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: md2tp <file_path>")
		os.Exit(1)
	}
	// Read the markdown file
	markdownFile := os.Args[1]

	if !fileExists(markdownFile) {
		log.Printf("File does not exist: %s\n", markdownFile)
		os.Exit(1)
	}

	markdownContent, err := os.ReadFile(markdownFile)
	if err != nil {
		log.Printf("Failed to read markdown file: %v", err)
		os.Exit(2)
	}

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
	err = chromedp.Run(ctx, generatePDF(uri, &pdfBuffer))
	if err != nil {
		fmt.Printf("Failed to generate file: %v \n", uri)
		log.Fatalf("Failed to generate PDF: %v\n", err)
	}

	// Write the PDF to a file
	// Strip extension from markdown file

	pdfFile := stripExtension(markdownFile)+".pdf"
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

func generatePDF(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
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
