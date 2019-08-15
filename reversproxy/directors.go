package reversproxy

import (
	"net/http"
	"net/url"
)

// NewPlusPathDirector - director, который контструирует out-запрос из пркоси и добавляет к path общий префикс админ-панели.
// Напрмиер, "/admin/ ..."
func NewPlusPathDirector(targetUrl *url.URL, resource string) (director func(r *http.Request)) {

	director = func(req *http.Request) {

		req.URL.Scheme = targetUrl.Scheme

		req.Host = targetUrl.Host
		req.URL.Host = targetUrl.Host

		req.URL.Path = targetUrl.Path + resource
		req.RequestURI = targetUrl.Path + resource
	}

	return
}

// NewResourceDirector - director, который контструирует out-запрос из пркоси,
// прибавляя к path только запрашиваемы ресурс.
func NewResourceDirector(targetUrl *url.URL, resource string) (director func(r *http.Request)) {

	director = func(req *http.Request) {

		req.URL.Scheme = targetUrl.Scheme

		req.Host = targetUrl.Host
		req.URL.Host = targetUrl.Host

		req.URL.Path = resource
		req.RequestURI = resource
	}

	return
}
