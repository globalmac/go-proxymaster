package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var urls = []string{
	"https://check-host.net/ip",
	"https://check-host.net/ip",
	"https://check-host.net/ip",
}

type HttpResponse struct {
	url      string
	response *http.Response
	err      error
}

func asyncHttpGets(urls []string) []*HttpResponse {

	ch := make(chan *HttpResponse)
	responses := make([]*HttpResponse, 0)

	// Proxy like - "//1.1.1.1:2222"
	proxies := []string{
		"//PROXY_URL:PROXY_PORT",
		"//PROXY_URL:PROXY_PORT",
		"//PROXY_URL:PROXY_PORT",
	}

	for _, ul := range urls {

		// For tests
		//time.Sleep(200 * time.Millisecond)
		
		// Make some random proxy for URL
		rand.Seed(time.Now().UnixNano())
		randURL := proxies[rand.Intn(len(proxies)-1)]
		proxyUrl, err := url.Parse(randURL)

		if err != nil {
			panic(err)
		}

		go func(ul string) {

			var netTransport = &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
				TLSHandshakeTimeout: 10 * time.Second,
			}

			var client = &http.Client{
				Timeout:   time.Second * 10,
				Transport: netTransport,
			}

			fmt.Printf("Fetching %s \n", ul)

			resp, err := client.Get(ul)

			ch <- &HttpResponse{ul, resp, err}

			if err != nil && resp != nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
			}

		}(ul)
	}

	for {
		select {
		case r := <-ch:
			fmt.Printf("%s was fetched\n", r.url)
			if r.err != nil {
				fmt.Println("with an error", r.err)
			}
			responses = append(responses, r)
			if len(responses) == len(urls) {
				return responses
			}
		case <-time.After(50 * time.Millisecond):
			fmt.Printf(".")
		}
	}

}

func main() {
	results := asyncHttpGets(urls)
	for _, result := range results {
		if result != nil && result.response != nil {
			reader := bufio.NewReader(result.response.Body)
			data, _, _ := reader.ReadLine()
			fmt.Println(result.response.Status + " => " + string(data))
		}
	}
}
