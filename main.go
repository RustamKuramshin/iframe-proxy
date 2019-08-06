package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

var targetUrl string

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/pageproxy", pageProxyHandle).Methods("GET")
	router.HandleFunc("/resourceproxy", resourceProxyHandle).Methods("GET")
	router.HandleFunc("/proxy", proxyHandle).Methods("GET")

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	log.Println("Server starting")
	// log.Fatal(http.ListenAndServeTLS(":9090", "cert.pem", "key.pem", router))
	log.Fatal(http.ListenAndServe(":9090", router))
}

func pageProxyHandle(res http.ResponseWriter, req *http.Request) {

	queryParams := req.URL.Query()
	targetUrl = queryParams["url"][0]
	proxyUrl, _ := url.Parse(queryParams["url"][0])

	proxy := NewReversProxy()
	director := func(req *http.Request) {
		req.URL.Scheme = proxyUrl.Scheme
		req.URL.Host = proxyUrl.Host
		req.URL.Path = proxyUrl.Path
	}
	proxy.Director = director
	proxy.ModifyResponse = func(proxyRes *http.Response) error {

		proxyRes.Header.Del("X-Frame-Options")

		defer proxyRes.Body.Close()

		doc, err := goquery.NewDocumentFromReader(proxyRes.Body)

		if err == nil {
			doc.Find("link").Each(func(i int, s *goquery.Selection) {
				if valHref, exst := s.Attr("href"); exst {
					s.SetAttr("href", "/resourceproxy?url="+targetUrl+valHref)
				}
			})
			doc.Find("script").Each(func(i int, s *goquery.Selection) {
				if valHref, exst := s.Attr("src"); exst {
					s.SetAttr("src", "/resourceproxy?url="+targetUrl+valHref)
				}
			})
		}

		html, _ := doc.Html()
		bodyByte := []byte(html)
		proxyRes.Body = ioutil.NopCloser(bytes.NewReader(bodyByte))
		proxyRes.ContentLength = int64(len(bodyByte))
		proxyRes.Header.Set("Content-Length", strconv.Itoa(len(bodyByte)))
		go printResponse(proxyRes)
		return nil
	}

	proxy.ServeHTTP(res, req)
}

func resourceProxyHandle(res http.ResponseWriter, req *http.Request) {

	// TODO: js modifyzer

	queryParams := req.URL.Query()
	proxyUrl, _ := url.Parse(queryParams["url"][0])

	proxy := NewReversProxy()
	proxy.Director = MakeStandartDirector(req, proxyUrl)
	proxy.ModifyResponse = func(proxyRes *http.Response) error {
		proxyRes.Header.Del("X-Frame-Options")
		go printResponse(proxyRes)
		return nil
	}

	proxy.ServeHTTP(res, req)
}

func proxyHandle(res http.ResponseWriter, req *http.Request) {

	queryParams := req.URL.Query()
	proxyUrl, _ := url.Parse(queryParams["url"][0])

	proxy := NewReversProxy()
	proxy.Director = MakeStandartDirector(req, proxyUrl)
	proxy.ModifyResponse = func(proxyRes *http.Response) error {
		proxyRes.Header.Del("X-Frame-Options")
		printResponse(proxyRes)
		return nil
	}

	proxy.ServeHTTP(res, req)
}

func NewReversProxy() *httputil.ReverseProxy {

	return &httputil.ReverseProxy{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return http.ProxyFromEnvironment(req)
			},
			Dial: func(network, addr string) (net.Conn, error) {
				conn, err := (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 3 * time.Second,
				}).Dial(network, addr)
				if err != nil {
					println("Error during DIAL:", err.Error())
				}
				return conn, err
			},
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

}

func printResponse(res *http.Response) {
	log.Println("HTTP response")
	if body, err := httputil.DumpResponse(res, true); err == nil {
		fmt.Println(string(body))
	}
}

func MakeStandartDirector(httpreq *http.Request, proxyUrl *url.URL) func(r *http.Request) {
	return func(req *http.Request) {
		httpreq.Host = proxyUrl.Host
		httpreq.URL.Host = proxyUrl.Host
		httpreq.URL.Path = proxyUrl.Path
		httpreq.URL.Scheme = proxyUrl.Scheme
		httpreq.RequestURI = proxyUrl.Path
	}
}
