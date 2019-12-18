package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
)

// Doc defines a Yumpu document
type Doc struct {
	Document struct {
		url      string
		dir      string
		Title    string `json:"title"`
		BasePath string `json:"base_path"`
		Pages    []Page `json:"pages"`
	}
}

// Page defines a single page within a Yumpu document
type Page struct {
	Number int               `json:"nr"`
	Images map[string]string `json:"images"`
	Qss    map[string]string `json:"qss"`
}

func main() {
	var d Doc

	flagID := flag.String("id", "", "Yumpu document ID")
	flagOut := flag.String("out", ".", "directory where to download jpeg files")
	flag.Parse()

	if *flagID == "" { // flag not set
		log.Fatal("Yumpu document ID not set")
	}

	u := &url.URL{
		Scheme: "https",
		Host:   "www.yumpu.com",
		Path:   path.Join("document/json2/", *flagID),
	}

	d.Document.url = u.String()
	d.Document.dir = *flagOut

	log.Printf("downloading JSON response from '%s'", d.Document.url)
	res, err := http.Get(d.Document.url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(b, &d)

	var wg sync.WaitGroup
	for _, p := range d.Document.Pages {
		pURL := fmt.Sprintf("%s%s?%s", d.Document.BasePath, p.Images["large"], p.Qss["large"])

		wg.Add(1)

		go func(p Page) {
			// log errors and carry on
			res, err := http.Get(pURL)
			if err != nil {
				log.Println(err)
			}
			defer res.Body.Close()

			pth := path.Join(d.Document.dir, fmt.Sprintf("%d.jpg", p.Number))
			f, err := os.Create(pth)
			if err != nil {
				log.Fatal(err)
			}

			defer f.Close()

			// dump the response body to the file
			_, err = io.Copy(f, res.Body)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("downloaded page %d", p.Number)
			wg.Done()
		}(p)
	}
	wg.Wait()
}
