package main

import (
	"fmt"

	"github.com/ThejasNU/sitemap-crawler/helpers"
	myTypes "github.com/ThejasNU/sitemap-crawler/types"
)

func main() {
	parser := myTypes.DefaultParser{}
	results := helpers.ScrapeSitemap("https://www.quicksprout.com/sitemap.xml",parser,10)

	for _, result := range results {
		fmt.Println(result)
	}

}
