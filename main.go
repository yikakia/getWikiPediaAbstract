package main

import (
	"encoding/xml"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Document 是用于写入文件的
type Document struct {
	Title string `xml:"title"`
	URL   string `xml:"url"`
	Text  string `xml:"abstract"`
}

// OutputForm 是写入文件的最高级
type OutputForm struct {
	XMLName xml.Name `xml:"doc"`
	Document
}

// PathExists 用于检测指定的 path 是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func loadfile(filename string) *os.File {
	_, _ = os.OpenFile(filename, os.O_CREATE, 0666)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func writefile(filename string, doc OutputForm) {
	output, err := xml.MarshalIndent(doc, " ", "  ")
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	file := loadfile(filename)
	defer file.Close()
	file.Write(output)
	file.Write([]byte("\n"))
	// ioutil.WriteFile(filename, output, os.ModeAppend)
}
func main() {
	filename := ".\\test.xml"
	pagenums := 500
	waittime := time.Duration(time.Millisecond * 100)
	website := "https://en.wikipedia.org/wiki/Special:Random"
	proxyAddr := "http://localhost:10807"
	var finalurl string
	proxy, err := url.Parse(proxyAddr)

	if err != nil {
		log.Fatal(err)
	}
	netTransport := &http.Transport{
		Proxy:                 http.ProxyURL(proxy),
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: time.Second * time.Duration(5),
	}
	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			finalurl = "https://" + req.URL.Host + req.URL.Path
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
	for i := 0; i < pagenums; i++ {
		response, err := httpClient.Get(website)
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()

		doc, err := goquery.NewDocumentFromResponse(response)
		if err != nil {
			log.Fatal(err)
		}

		var mydoc = OutputForm{}
		mydoc.URL = finalurl
		doc.Find("body").Find("h1[id=firstHeading]").Each(func(i int, selection *goquery.Selection) {
			mydoc.Title = selection.Text()
		})
		doc.Find("body").Find("div[id=bodyContent]").Find("div[class=mw-parser-output]").Find("p:first-of-type").Each(func(i int, selection *goquery.Selection) {
			mydoc.Text = selection.Text()
		})
		if mydoc.Text == "\n" || mydoc.Text == "\n\n" {
			doc.Find("body").Find("div[id=bodyContent]").Find("div[class=mw-parser-output]").Find("p:first-of-type").Next().Each(func(i int, selection *goquery.Selection) {
				mydoc.Text += selection.Text()
			})
		}
		// doc.Find("body").Find("div[id=bodyContent]").Find("div[class=mw-parser-output]").Find("p:first-of-type").Next().Each(func(i int, selection *goquery.Selection) {
		// 	mydoc.Text += selection.Text()
		// })
		mydoc.Title = strings.Replace(mydoc.Title, "\n", "", -1)
		mydoc.Text = strings.Replace(mydoc.Text, "\n", "", -1)
		mydoc.URL = strings.Replace(mydoc.URL, "\n", "", -1)
		if len(mydoc.Title)*2 > len(mydoc.Text) {
			log.Printf("too short in %s of %s", mydoc.Title, mydoc.Text)
			continue
		}
		writefile(filename, mydoc)
		time.Sleep(waittime)
		log.Printf("file %d is writen,finesed %f %% of %d files  \n", i, float32(i)/float32(pagenums), pagenums)
	}

}
