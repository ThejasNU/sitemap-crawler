package helpers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	myTypes "github.com/ThejasNU/sitemap-crawler/types"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

func randomUserAgent() string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(userAgents)
	return userAgents[randNum]
}

//extract all the URLs in a website
func extractURLs(resp *http.Response) ([]string,error){
	doc,err := goquery.NewDocumentFromResponse(resp)
	if err!=nil{
		return nil,err
	}

	results := []string{}
	sel := doc.Find("loc")

	for i := range sel.Nodes{
		loc := sel.Eq(i)
		result := loc.Text()
		results = append(results, result)
	}
	return results,nil
}

//makes request to get content of a webpage 
func makeRequest(url string) (*http.Response,error) {
	client := http.Client{
		Timeout: 10*time.Second,
	}
	
	req,err := http.NewRequest("GET",url,nil)
	req.Header.Set("User-Agent",randomUserAgent())

	if err!=nil{
		return nil,err
	}

	res,err := client.Do(req)
	if err!=nil{
		return nil,err
	}

	return res,nil
}

//seperate websites which have sitemap and which do not
func isSitemap(urls []string) ([]string,[]string){
	sitemapFiles := []string{}
	pages := []string{}

	for _,page := range urls{
		foundSitemap := strings.Contains(page,"xml")
		if foundSitemap==true{
			fmt.Println("Found sitemap",page)
			sitemapFiles = append(sitemapFiles, page)
		} else{
			pages= append(pages,page)
		}
	}
	return sitemapFiles,pages
}

//extract pages which don't have sitemap
func extractSitemapURLs(startURL string) []string {
	worklist := make(chan []string)
	toCrawl := []string{}

	go func() { worklist <- []string{startURL} }()

	//loop which goes on till we have something in channel
	var n int = 1
	for ; n > 0; n-- {
		listOfLinks := <-worklist
		for _, link := range listOfLinks {
			n++
			go func(link string) {
				response, err := makeRequest(link)
				if err != nil {
					log.Printf("Error retrieving URL: %s", link)
				}

				urls, err := extractURLs(response)
				if err != nil {
					log.Printf("Error extracting document from response of URL: %s", link)
				}

				sitemapFiles, pages := isSitemap(urls)
				if sitemapFiles != nil {
					worklist <- sitemapFiles
				}
				toCrawl = append(toCrawl, pages...)

			}(link)
		}
	}
	return toCrawl

}

func scrapeURLs(urls []string,parser myTypes.Parser,concurrency int) []myTypes.SeoData {
	tokens := make(chan struct{},concurrency)
	worklist := make(chan []string)
	results := []myTypes.SeoData{}
	
	go func() { worklist <- urls }()
	
	var n int = 1
	for ; n > 0; n-- {
		list := <-worklist
		for _, url := range list {
			if url != "" {
				n++
				go func(url string, token chan struct{}) {
					log.Printf("Requesting URL: %s", url)
					res, err := scrapePage(url, tokens, parser)
					if err != nil {
						log.Printf("Encountered error, URL: %s", url)
					} else {
						results = append(results, res)
					}
					worklist <- []string{}
				}(url, tokens)
			}
		}
	}
	return results
}

func scrapePage(url string,tokens chan struct{},parser myTypes.Parser) (myTypes.SeoData,error) {
	crawledRes,err := crawlPage(url,tokens)
	if err!=nil{
		return myTypes.SeoData{},err
	}

	seoData,err := parser.GetSeoData(crawledRes)
	if err!= nil{
		return myTypes.SeoData{},err
	}

	return seoData,nil
}

//crawl pages which do not have sitemaps
func crawlPage(url string,tokens chan struct{}) (*http.Response,error) {
	tokens <- struct{}{}
	resp, err := makeRequest(url)
	<-tokens
	if err != nil {
		return nil, err
	}
	return resp, err
}

func ScrapeSitemap(url string,parser myTypes.Parser,concurrency int) []myTypes.SeoData {
	sitemapURLs := extractSitemapURLs(url)
	scrapedData := scrapeURLs(sitemapURLs,parser,concurrency)
	return scrapedData
}
