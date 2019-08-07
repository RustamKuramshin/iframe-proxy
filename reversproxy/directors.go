package reversproxy

import (
	"net/http"
	"net/url"
)

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

func NewResourcesDirector(proxyUrl *url.URL, resPath string) func(r *http.Request) {

	var resourcesDirector = func(req *http.Request) {
		req.Host = proxyUrl.Host
		req.URL.Host = proxyUrl.Host
		req.URL.Path = proxyUrl.Path + resPath
		req.URL.Scheme = proxyUrl.Scheme
		req.RequestURI = proxyUrl.Path + resPath
	}

	return resourcesDirector

}

func NewXhrDirector(proxyUrl *url.URL) func(r *http.Request) {

	var resourcesDirector = func(req *http.Request) {
		req.Host = proxyUrl.Host
		req.URL.Host = proxyUrl.Host
		req.URL.Scheme = proxyUrl.Scheme
	}

	return resourcesDirector

}
