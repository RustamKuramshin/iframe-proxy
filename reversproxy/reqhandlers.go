package reversproxy

import (
	"net/http"
	"net/url"
	"strings"
)

func PageProxyHandle(res http.ResponseWriter, req *http.Request) {
	targetUrlStr := req.URL.Query()["url"][0]
	proxy := NewReversProxyFromTargetUrl(getTargetUrlFromUrlString(targetUrlStr))
	proxy.ModifyResponse = NewModifyResponseOverwriteRelPaths(targetUrlStr)
	proxy.ServeHTTP(res, req)
}

func TransparentProxyHandle(res http.ResponseWriter, req *http.Request) {
	proxy := NewReversProxyFromTargetUrl(getTargetUrlFromUrlString(req.URL.Query()["url"][0]))
	proxy.ModifyResponse = NewModifyResponseCutXFrame()
	proxy.ServeHTTP(res, req)
}

func TransparentXhrProxyHandle(res http.ResponseWriter, req *http.Request) {

	resPath := strings.Replace(req.RequestURI, "/transparentxhrproxy", "", 1)
	urlFromHeader := req.Header.Get("Referer")
	if targetUrl, err := url.Parse(urlFromHeader); err == nil {
		if urlQueryParam, ok := targetUrl.Query()["url"]; ok {
			proxy := NewReversProxyFromTargetUrl(getTargetUrlFromUrlString(urlQueryParam[0]))
			proxy.Director = NewTransparentXhrDirector(getTargetUrlFromUrlString(urlQueryParam[0]), resPath)
			proxy.ServeHTTP(res, req)
		}
	}
}

func IframeProxyHandle(res http.ResponseWriter, req *http.Request) {

	if urlQueryParam, ok := req.URL.Query()["url"]; ok {
		proxy := NewReversProxyFromTargetUrl(getTargetUrlFromUrlString(urlQueryParam[0]))
		proxy.Director = NewStandartDirector(getTargetUrlFromUrlString(urlQueryParam[0]))
		proxy.ModifyResponse = NewModifyResponseChangeXhrBehavior()
		proxy.ServeHTTP(res, req)
	} else {
		resPath := strings.Replace(req.RequestURI, "/iframeproxy/", "", 1)
		urlFromHeader := req.Header.Get("Referer")
		if targetUrl, err := url.Parse(urlFromHeader); err == nil {
			if urlQueryParam, ok := targetUrl.Query()["url"]; ok {
				proxy := NewReversProxyFromTargetUrl(getTargetUrlFromUrlString(urlQueryParam[0]))
				proxy.Director = NewResourcesDirector(getTargetUrlFromUrlString(urlQueryParam[0]), resPath)
				proxy.ServeHTTP(res, req)
			}
		}
	}
}

func XhrProxyHandle(res http.ResponseWriter, req *http.Request) {

	urlFromHeader := req.Header.Get("Referer")
	resPath := strings.Replace(req.RequestURI, "/xhrproxy", "", 1)

	if targetUrl, err := url.Parse(urlFromHeader); err == nil {
		if urlQueryParam, ok := targetUrl.Query()["url"]; ok {
			proxy := NewReversProxyFromTargetUrl(getTargetUrlFromUrlString(urlQueryParam[0]))
			proxy.Director = NewXhrDirector(getTargetUrlFromUrlString(urlQueryParam[0]), resPath)
			proxy.ServeHTTP(res, req)
		}
	}
}

func getTargetUrlFromUrlString(urlStr string) *url.URL {
	targetUrl, _ := url.Parse(urlStr)
	return targetUrl
}
