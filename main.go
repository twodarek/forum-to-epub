package main

import (
	"bufio"
	"flag"
	"github.com/bmaupin/go-epub"
	"github.com/twodarek/go-htmlutil"
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
	epubCSSFile := "assets/epub.css"
	preFontFile := "assets/SourceCodePro-Regular.ttf"

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
		log.Printf(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error: unable to read all lines from file: %s", err)
	}

	chapters := []epubChapter{}
	// Get the posts

	log.Printf("chapterTitlesAndLinks: %s", chapterTitlesAndLinks)

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
		log.Println(articleNode)
		log.Println(postAnchor)
		articleBodyWrapper := htmlutil.GetFirstHtmlNode(articleNode, "article", "class", "message-body")
		log.Println(articleBodyWrapper)
		articleBody := htmlutil.GetFirstHtmlNode(articleBodyWrapper, "div", "class", "bbWrapper")
		log.Println(articleBody)
		log.Printf("title: %s, text: %s", chapterLink[0], chapterLink[1])

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

	_, err = book.AddFont(preFontFile, "")
	if err != nil {
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
				log.Printf("Error in dumping html to string while adding a chapter to the book: %s", err)
			}
			chapterContent += nodeContent
		}

		log.Printf("chapterContent: %s", chapterContent)

		_, err := book.AddSection(chapterContent, chapter.title, chapter.filename, epubCSSPath)
		if err != nil {
			log.Printf("Error in adding a chapter to the book: %s", err)
		}
	}


	err = book.Write("theLastAngel.epub")
	if err != nil {
		log.Printf("Error in writing out the resulting file: %s", err)
	}

}
