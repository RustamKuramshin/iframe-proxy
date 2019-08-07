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

	targetUrlStr := req.URL.Query()["url"][0]
	proxyUrl := getTargetUrlFromUrlString(targetUrlStr)

	proxy := NewReversProxy()
	proxy.Director = NewStandartDirector(proxyUrl)
	proxy.ModifyResponse = NewModifyResponseOverwriteRelPaths(targetUrlStr)
	proxy.ServeHTTP(res, req)
}

func resourceProxyHandle(res http.ResponseWriter, req *http.Request) {

	// TODO: js modifyzer

	proxyUrl := getTargetUrlFromUrlString(req.URL.Query()["url"][0])

	proxy := NewReversProxy()
	proxy.Director = NewStandartDirector(proxyUrl)
	proxy.ModifyResponse = NewCutXFrame()
	proxy.ServeHTTP(res, req)
}

func proxyHandle(res http.ResponseWriter, req *http.Request) {

	proxyUrl := getTargetUrlFromUrlString(req.URL.Query()["url"][0])

	proxy := NewReversProxy()
	proxy.Director = NewStandartDirector(proxyUrl)
	proxy.ModifyResponse = NewCutXFrame()
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

func NewModifyResponseOverwriteRelPaths(targetUrl string) func(r *http.Response) error {

	var modifyResponseOverwriteRelPaths = func(res *http.Response) error {

		res.Header.Del("X-Frame-Options")

		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)

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
		res.Body = ioutil.NopCloser(bytes.NewReader(bodyByte))
		res.ContentLength = int64(len(bodyByte))
		res.Header.Set("Content-Length", strconv.Itoa(len(bodyByte)))
		printResponse(res)
		return nil
	}

	return modifyResponseOverwriteRelPaths

}

func NewCutXFrame() func(r *http.Response) error {

	var cutXFrame = func(res *http.Response) error {
		res.Header.Del("X-Frame-Options")
		printResponse(res)
		return nil
	}

	return cutXFrame
}

func NewStandartDirector(proxyUrl *url.URL) func(r *http.Request) {

	var standartDirector = func(req *http.Request) {
		req.Host = proxyUrl.Host
		req.URL.Host = proxyUrl.Host
		req.URL.Path = proxyUrl.Path
		req.URL.Scheme = proxyUrl.Scheme
		req.RequestURI = proxyUrl.Path
	}

	return standartDirector

}

func getTargetUrlFromUrlString(urlStr string) *url.URL {
	targetUrl, _ := url.Parse(urlStr)
	return targetUrl
}
