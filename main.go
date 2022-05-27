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
	inputFilePath := flag.String("input-file", "", "The path to file containing pages that you want to include")
	titleIn := flag.String("title", "The Last Angel", "The title of the book")
	authorIn := flag.String("author", "Proximal Flame", "The author of the book")
	destFileIn := flag.String("output-file", "theLastAngel.pub", "The location of the intended output file")
	flag.Parse()

	if *inputFilePath == "" {
		log.Fatalf("Error: input-file required, but not found")
	}

	title := *(titleIn)
	author := *(authorIn)
	destFile := *(destFileIn)

	epubCSSFile := "assets/epub.css"
	preFontFile := "assets/SourceCodePro-Regular.ttf"

	book := epub.NewEpub(title)
	book.SetAuthor(author)

	inputFile, err := os.Open(*inputFilePath)
	if err != nil {
		log.Fatalf("Error: unable to open input file: %s", err)
	}
	defer inputFile.Close()

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	chapterTitlesAndLinks := [][]string{}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")
		if !strings.Contains(line[0], "#") {
			resp, err := httpClient.Get(line[0])
			if err != nil {
				log.Fatalf("Error: unable to resolve actual url of post: %s, %s", line[0], err)
			}
			newUrl := resp.Header.Get("location")
			if newUrl == "" {
				log.Fatalf("Error: unable to resolve actual url of post, location header empty/nonexistant: %s, %s", line[0], err)
			}
			line[0] = newUrl
		}
		chapterTitlesAndLinks = append(chapterTitlesAndLinks, line)
		log.Printf(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error: unable to read all lines from file: %s", err)
	}

	chapters := []epubChapter{}
	// Get the posts

	log.Printf("chapterTitlesAndLinks: %s", chapterTitlesAndLinks)

	for count, chapterLink := range chapterTitlesAndLinks {
		resp, err := http.Get(chapterLink[0])
		if err != nil {
			log.Printf("Error: Unable to get link %d, %s, because of error %s", count, chapterLink[0], err)
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

		postAnchor := strings.Split(chapterLink[0], "#")[1]
		articleNode := htmlutil.GetFirstHtmlNode(doc, "article", "data-content", postAnchor)
		articleBody := htmlutil.GetFirstHtmlNode(articleNode, "div", "class", "bbWrapper")
		log.Printf("title: %s, text: %s", chapterLink[1], chapterLink[0])

		chapterArray := []html.Node{*articleBody}

		chapter := epubChapter{
			title:    chapterLink[1],
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

		_, err := book.AddSection(chapterContent, chapter.title, chapter.filename, epubCSSPath)
		if err != nil {
			log.Printf("Error in adding a chapter to the book: %s", err)
		}
	}


	err = book.Write(destFile)
	if err != nil {
		log.Printf("Error in writing out the resulting file: %s", err)
	}

}
