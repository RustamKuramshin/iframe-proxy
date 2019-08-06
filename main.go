package main

import (
	"bytes"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var targetUrl string
var proxySrcUrl string

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/proxy", proxyHandle).Methods("GET")
	router.HandleFunc("/srcproxy", srcproxyHandle).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	log.Println("Server starting")
	// log.Fatal(http.ListenAndServeTLS(":9090", "cert.pem", "key.pem", router))
	log.Fatal(http.ListenAndServe(":9090", router))
}

func proxyHandle(res http.ResponseWriter, req *http.Request) {

	queryParams := req.URL.Query()
	proxyUrl, _ := url.Parse(queryParams["url"][0])
	targetUrl := queryParams["url"][0]

	director := func(req *http.Request) {
		req.URL.Scheme = proxyUrl.Scheme
		req.URL.Host = proxyUrl.Host
		req.URL.Path = proxyUrl.Path
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
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

	proxy.ModifyResponse = func(proxyRes *http.Response) error {

		proxyRes.Header.Del("X-Frame-Options")

		defer proxyRes.Body.Close()

		doc, err := goquery.NewDocumentFromReader(proxyRes.Body)

		if err == nil {
			doc.Find("link").Each(func(i int, s *goquery.Selection) {
				if valHref, exst := s.Attr("href"); exst {
					s.SetAttr("href", "http://localhost:9090/srcproxy?url="+targetUrl+valHref)
				}
			})
			doc.Find("script").Each(func(i int, s *goquery.Selection) {
				if valHref, exst := s.Attr("src"); exst {
					s.SetAttr("src", "http://localhost:9090/srcproxy?url="+targetUrl+valHref)
				}
			})
		}

		html, _ := doc.Html()
		bodyByte := []byte(html)
		proxyRes.Body = ioutil.NopCloser(bytes.NewReader(bodyByte))

		proxyRes.ContentLength = int64(len(bodyByte))
		proxyRes.Header.Set("Content-Length", strconv.Itoa(len(bodyByte)))

		log.Println("Server response received")
		//if body, err := httputil.DumpResponse(proxyRes, true); err == nil {
		//	//fmt.Println(string(body))
		//}
		return nil
	}

	proxy.ServeHTTP(res, req)
}

func srcproxyHandle(res http.ResponseWriter, req *http.Request) {

	queryParams := req.URL.Query()
	proxyUrl, _ := url.Parse(queryParams["url"][0])

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: proxyUrl.Scheme,
		Host:   proxyUrl.Host,
	})

	req.Host = proxyUrl.Host
	req.URL.Host = proxyUrl.Host
	req.URL.Path = proxyUrl.Path
	req.URL.Scheme = proxyUrl.Scheme
	req.RequestURI = proxyUrl.Path

	proxy.ModifyResponse = func(proxyRes *http.Response) error {
		return nil
	}

	proxy.ServeHTTP(res, req)
}
