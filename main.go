package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)
var MaxDepth = 2
var wg sync.WaitGroup
var linksresult []Link

type Link struct{
	text 	string
	depth 	int
	url 	string
}
func(self Link) Valid() bool{
	if self.depth>=MaxDepth{
		return false
	}
	if len(self.text)==0{
		return false
	}
	if len(self.url) == 0{
		return false
	}
	return true
}

//returns a slice of links gathered from a given html file
func LinkReader(r *http.Response, depth int) []Link{
	page := html.NewTokenizer(r.Body)
	var links []Link

	var start *html.Token
	var text string
	for {
		page.Next()
		token:=page.Token()
		if token.Type==html.ErrorToken{
			break
		}
		if start!=nil && token.Type==html.TextToken{
			text=fmt.Sprintf("%s%s", text, token.Data)
		}
		switch token.Type{
		case html.StartTagToken:
			if len(token.Attr) > 0 {
				start = &token
			}
		case html.EndTagToken:
			if start == nil {
				continue
			}
			link := NewLink(*start, text, depth)
			if link.Valid() {
				fmt.Printf("Link Found %v", link.url)
				links = append(links, link)
			}

			start = nil
			text = ""
			
		}
		
	}
	return links
}

func NewLink(tag html.Token, text string, depth int) Link{
	link:=Link{text:strings.TrimSpace(text),depth:depth}
	for i:= range tag.Attr{
		if tag.Attr[i].Key=="href"{
			link.url=strings.TrimSpace(tag.Attr[i].Val)
		}
	}
	return link
}

func GetLinksFromURL(url string, depth int, linkschan chan Link){
	defer wg.Done()
	page, err := GetResponseFromURL(url)
	if err != nil {
		fmt.Print(err)
		return 
	}
	links := LinkReader(page, depth)
	if depth>MaxDepth{
		return 
	}
	for i:=range links{
		linkschan<-links[i]
	}
	return 
}


func GetResponseFromURL(url string) (resp *http.Response, err error) {
	resp, err = http.Get(url)
	if err != nil {
		fmt.Printf("Error: %s \n", err)
		return
	}

	if resp.StatusCode > 299 {
		return
	}
	return

}
//var maxWorkers = 10

func start(links []Link, depth int, linkschan chan Link){
	if depth>MaxDepth{
		return
	}
	for i := range links{
		
		wg.Add(1)
		go GetLinksFromURL(links[i].url, depth+1, linkschan)
	}
	start (linksresult, depth+1, linkschan)
}
func main(){
	
	wg.Add(1)
	linkschan:=make(chan Link)
	go func(linkschan chan Link){
		for{
			select {
			case msg:= <-linkschan:
				linksresult=append(linksresult, msg)
				fmt.Print(msg)
				fmt.Print("\n")
			}
		}
	}(linkschan)
	GetLinksFromURL(os.Args[1], 0, linkschan)
	start(linksresult, 0, linkschan)
	wg.Wait()
	fmt.Print(len(linkschan))
	fmt.Print(" ")
	fmt.Print(len(linksresult))
}