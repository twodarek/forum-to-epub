package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bmaupin/go-epub"
	"github.com/bmaupin/go-htmlutil"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"os"
	"strings"
)

type epubChapter struct {
	title    string
	filename string
	nodes    []html.Node
}

func main() {
	log.SetFlags(0)
	inputFilePath := flag.String("input-file", "", "path to file containing pages that you want to include")
	flag.Parse()

	if *inputFilePath == "" {
		log.Fatalf("Error: input-file required, but not found")
	}

	title := "The Last Angel"
	author := "Proximal Flame"
	epubCSSFile := "assets/styles/epub.css"

	book := epub.NewEpub(title)
	book.SetAuthor(author)

	inputFile, err := os.Open(*inputFilePath)
	if err != nil {
		log.Fatalf("Error: unable to open input file: %s", err)
	}
	defer inputFile.Close()

	chapterTitlesAndLinks := [][]string{}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		chapterTitlesAndLinks = append(chapterTitlesAndLinks, strings.Split(scanner.Text(), ","))
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error: unable to read all lines from file: %s", err)
	}

	chapters := []epubChapter{}
	// Get the posts

	for count, chapterLink := range chapterTitlesAndLinks {
		resp, err := http.Get(chapterLink[1])
		if err != nil {
			log.Printf("Error: Unable to get link %d, %s, because of error %s", count, chapterLink[1], err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()

		doc, err := html.Parse(resp.Body)
		if err != nil {
			log.Fatalf("Parse error: %s", err)
		}

		postAnchor := strings.Split(chapterLink[1], "#")[1]
		articleNode := htmlutil.GetFirstHtmlNode(doc, "article", "data-content", postAnchor)
		articleBodyWrapper := htmlutil.GetFirstHtmlNode(articleNode, "article", "class", "message-body")
		articleBody := htmlutil.GetFirstHtmlNode(articleBodyWrapper, "div", "class", "bbWrapper")

		chapterArray := []html.Node{*articleBody}

		chapter := epubChapter{
			title:    chapterLink[0],
			filename: "",
			nodes:    chapterArray,
		}
		chapters = append(chapters, chapter)
	}

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


	err = book.Write("thwLastAngel.epub")
	if err != nil {
		log.Printf("Error in writing out the resulting file: %s", err)
	}

}
