package reversproxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func NewReversProxyFromTargetUrl(proxyUrl *url.URL) *httputil.ReverseProxy {

	return &httputil.ReverseProxy{
		Director: NewStandartDirector(proxyUrl),
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

func PrintResponse(res *http.Response) {
	log.Println("HTTP response")
	if body, err := httputil.DumpResponse(res, true); err == nil {
		fmt.Println(string(body))
	}
}
