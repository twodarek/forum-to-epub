package main

import (
	"github.com/bmaupin/go-epub"
	"github.com/bmaupin/go-htmlutil"
	"golang.org/x/net/html"
	"log"
	"strings"
)

type epubChapter struct {
	title    string
	filename string
	nodes    []html.Node
}

func main() {

	title := "The Last Angel"
	author := "Proximal Flame"
	epubCSSFile := "assets/styles/epub.css"

	book := epub.NewEpub(title)
	book.SetAuthor(author)

	// Get the posts

	chapters := []epubChapter{}
	// Write to epub

	epubCSSPath, err := book.AddCSS(epubCSSFile, "")
	if err != nil{
		log.Printf("Error occurred while attempting to add css: %s", err)
	}

	// Iterate through each chapter and add it to the EPUB
	internalLinks := make(map[string]string)

	for _, chapter := range chapters {
		chapterContent := ""
		for _, chapterNode := range chapter.nodes {
			// Fix internal links so they work after splitting page into chapters
			for _, linkNode := range htmlutil.GetAllHtmlNodes(&chapterNode, "a", "", "") {
				for i, attr := range linkNode.Attr {
					if attr.Key == "href" && strings.HasPrefix(attr.Val, "#") {
						linkNode.Attr[i].Val = internalLinks[attr.Val[1:]]
					}
				}
			}

			nodeContent, err := htmlutil.HtmlNodeToString(&chapterNode)
			if err != nil {
				log.Printf("Error in adding a chapter to the book: %s", err)
			}
			chapterContent += nodeContent
		}

		_, err := book.AddSection(chapterContent, chapter.title, chapter.filename, epubCSSPath)
		if err != nil {
			log.Printf("Error in adding a chapter to the book: %s", err)
		}
	}


	err := book.Write("thwLastAngel.epub")
	if err != nil {
		log.Printf("Error in writing out the resulting file: %s", err)
	}

}
