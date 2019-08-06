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
var proxySrcUrl string

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/pageproxy", pageProxyHandle).Methods("GET")
	router.HandleFunc("/resourceproxy", resourceProxyHandle).Methods("GET")
	router.HandleFunc("/proxy", proxyHandle).Methods("GET")

	router.PathPrefix("/js").HandlerFunc(proxyRootHandle)
	router.PathPrefix("/css").HandlerFunc(proxyRootHandle)
	router.PathPrefix("/meta").HandlerFunc(proxyRootHandle)
	router.PathPrefix("/controllers").HandlerFunc(proxyRootHandle)
	router.PathPrefix("/third-party").HandlerFunc(proxyRootHandle)
	router.PathPrefix("/modules").HandlerFunc(proxyRootHandle)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	log.Println("Server starting")
	// log.Fatal(http.ListenAndServeTLS(":9090", "cert.pem", "key.pem", router))
	log.Fatal(http.ListenAndServe(":9090", router))
}

func pageProxyHandle(res http.ResponseWriter, req *http.Request) {

	queryParams := req.URL.Query()
	targetUrl = queryParams["url"][0]
	proxyUrl, _ := url.Parse(queryParams["url"][0])

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

		log.Println("Return to client")
		if body, err := httputil.DumpResponse(proxyRes, true); err == nil {
			fmt.Println(string(body))
		}

		return nil
	}

	proxy.ServeHTTP(res, req)
}

func resourceProxyHandle(res http.ResponseWriter, req *http.Request) {

	// TODO: js modifyzer

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
		proxyRes.Header.Del("X-Frame-Options")
		log.Println("Return to client")
		if body, err := httputil.DumpResponse(proxyRes, true); err == nil {
			fmt.Println(string(body))
		}
		return nil
	}

	proxy.ServeHTTP(res, req)
}

func proxyHandle(res http.ResponseWriter, req *http.Request) {

	queryParams := req.URL.Query()
	targetUrl = queryParams["url"][0]
	fmt.Println(targetUrl)
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
		proxyRes.Header.Del("X-Frame-Options")
		log.Println("Return to client")
		if body, err := httputil.DumpResponse(proxyRes, true); err == nil {
			fmt.Println(string(body))
		}
		return nil
	}

	proxy.ServeHTTP(res, req)
}

func proxyRootHandle(res http.ResponseWriter, req *http.Request) {

	proxyUrl, _ := url.Parse(targetUrl)

	director := func(req *http.Request) {
		req.Host = proxyUrl.Host
		req.URL.Host = proxyUrl.Host
		req.URL.Path = proxyUrl.Path + req.URL.Path
		req.URL.Scheme = proxyUrl.Scheme
		req.RequestURI = proxyUrl.Path + req.RequestURI
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
		log.Println("Return to client")
		if body, err := httputil.DumpResponse(proxyRes, true); err == nil {
			fmt.Println(string(body))
		}
		return nil
	}

	proxy.ServeHTTP(res, req)
}
