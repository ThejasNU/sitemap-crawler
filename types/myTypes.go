package myTypes

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type SeoData struct {
	URL             string
	Title           string
	H1              string
	MetaDescription string
	StatusCode      int
}

type Parser interface {
	GetSeoData(resp *http.Response) (SeoData,error)
}

type DefaultParser struct {
}

func (parser DefaultParser) GetSeoData(resp *http.Response) (SeoData,error) {
	doc,err:= goquery.NewDocumentFromResponse(resp)
	if err!= nil{
		return SeoData{},err
	}

	metaData,_ := doc.Find("meta[name^=description]").Attr("content")
	result := SeoData{
		URL: resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Title: doc.Find("title").First().Text(),
		H1: doc.Find("h1").First().Text(),
		MetaDescription: metaData,
	}
	
	return result,nil
}
