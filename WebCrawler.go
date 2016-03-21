package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type FindUrl struct {
	v map[string]bool
	mux sync.Mutex 
}

func check(url string, findUrl FindUrl) bool {
	findUrl.mux.Lock()
	_, ok := findUrl.v[url]
	defer findUrl.mux.Unlock()	
	if (!ok) {
		findUrl.v[url] = true
		return true
	} else {
		return false
	}
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	findUrl := FindUrl{v: make(map[string]bool)}
	
	var wg sync.WaitGroup
	
	var checkUrl = func (url string, findUrlMap FindUrl) bool {
		findUrlMap.mux.Lock()
		_, ok := findUrlMap.v[url]
		defer findUrlMap.mux.Unlock()	
		if (!ok) {
			findUrlMap.v[url] = true
			return true
		} else {
			return false
		}
	}
	var crawl func(string, int)
	crawl = func(url string, depth int) {
		defer wg.Done()
		if depth <= 0 {
			return
		}
		
		if !checkUrl(url, findUrl) {
			return
		}
		
		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("found: %s %q\n", url, body)

		for _, u := range urls {
			wg.Add(1)
			go crawl(u, depth-1)
		}
	}
	
	wg.Add(1)
	crawl(url, depth)
	wg.Wait()
	
	return
}

func main() {

	Crawl("http://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
