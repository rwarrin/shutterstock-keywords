package main

import (
	"flag"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"fmt"

	"bufio"

	"golang.org/x/net/html"
)

const baseURL = "https://www.shutterstock.com/search/ids/"

func main() {
	// One shutterstock image ID per line
	inputPtr := flag.String("i", "ssids.txt", "the name and extension of the input file, defaults to ssids.txt")
	outputPtr := flag.String("o", "kws.csv", "the name and extension of the output file, defaults to kws.csv")
	flag.Parse()

	f1, err := os.Open(exPath() + "/" + *inputPtr)
	check(err)
	defer f1.Close()

	f2, err := os.OpenFile(exPath()+"/"+*outputPtr, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	check(err)
	defer f2.Close()

	scanner := bufio.NewScanner(f1)
	for scanner.Scan() {
		fmt.Print("Searching for image with ID: " + scanner.Text())
		resp, err := http.Get(baseURL + scanner.Text())
		check(err)
		if resp.Request.URL.RawQuery == "noid=1" {
			fmt.Print(" ...ERROR! invalid ID \n")
			_, err := f2.WriteString(scanner.Text() + ", " + "ERROR INVALID ID" + "\n")
			check(err)
			time.Sleep(time.Second * 5)
			return
		}
		fmt.Print(" ...SUCCESS! \n")
		kws := getKeywords(resp)
		_, err = f2.WriteString(scanner.Text() + ", " + kws + "\n")
		check(err)
		time.Sleep(time.Second * 5)
	}
}

// Simple function to check errors
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Simple function to get the path where the executable is located
func exPath() string {
	ex, err := os.Executable()
	check(err)
	return path.Dir(ex)
}

// Requests the HTML page based on the shutterstock ID passed. Parses the response
// body looking for the keywords and saving them into a slice. Finally joins
// the strings together into a single string of all keywords.
func getKeywords(resp *http.Response) string {
	z := html.NewTokenizer(resp.Body)
	defer resp.Body.Close()

	var kws []stringa

	fmt.Println("Building keywords list...")
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			fmt.Println("There was an error processing the HTML page")
			return "ERROR PROCESSING"

		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, v := range t.Attr {
					if v.Val == "pull-left btn btn-search-pill" {
						if tt = z.Next(); tt == html.TextToken {
							kws = append(kws, z.Token().Data)
						}
					}
				}
			}

		case tt == html.EndTagToken:
			t := z.Token()
			if t.Data == "html" {
				s := strings.Join(kws, ", ")
				fmt.Println("Savings keywords list...")
				return s
			}
		}
	}
}
