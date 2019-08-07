package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"proxyAdminPanels/reversproxy"
	"strings"
)

func main() {
	router := mux.NewRouter()

	//router.HandleFunc("/pageproxy", pageProxyHandle).Methods("GET")
	router.HandleFunc("/resourceproxy", transparentProxyHandle).Methods("GET")

	router.PathPrefix("/iframeproxy/").Methods("GET").HandlerFunc(iframeProxyHandle)
	router.PathPrefix("/").Headers("X-Requested-With", "XMLHttpRequest").HandlerFunc(xhrProxyHandle)
	router.PathPrefix("/").Headers("Accept", "application/json, text/plain, */*").HandlerFunc(xhrProxyHandle)
	router.PathPrefix("/").Headers("Accept", "text/event-stream").HandlerFunc(xhrProxyHandle)

	router.HandleFunc("/proxy", transparentProxyHandle).Methods("GET")

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(http.Dir("./web/"))))

	log.Println("Server starting")
	// log.Fatal(http.ListenAndServeTLS(":9090", "cert.pem", "key.pem", router))
	log.Fatal(http.ListenAndServe(":9090", router))
}

func pageProxyHandle(res http.ResponseWriter, req *http.Request) {
	targetUrlStr := req.URL.Query()["url"][0]
	proxy := reversproxy.NewReversProxyForTargetUrl(getTargetUrlFromUrlString(targetUrlStr))
	proxy.ModifyResponse = reversproxy.NewModifyResponseOverwriteRelPaths(targetUrlStr)
	proxy.ServeHTTP(res, req)
}

func transparentProxyHandle(res http.ResponseWriter, req *http.Request) {
	proxy := reversproxy.NewReversProxyForTargetUrl(getTargetUrlFromUrlString(req.URL.Query()["url"][0]))
	proxy.ModifyResponse = reversproxy.NewModifyResponseCutXFrame()
	proxy.ServeHTTP(res, req)
}

func iframeProxyHandle(res http.ResponseWriter, req *http.Request) {

	if urlQueryParam, ok := req.URL.Query()["url"]; ok {
		proxy := reversproxy.NewReversProxyForTargetUrl(getTargetUrlFromUrlString(urlQueryParam[0]))
		proxy.ServeHTTP(res, req)
	} else {
		resPath := strings.Replace(req.RequestURI, "/iframeproxy/", "", 1)
		urlFromHeader := req.Header.Get("Referer")
		if targetUrl, err := url.Parse(urlFromHeader); err == nil {
			if urlQueryParam, ok := targetUrl.Query()["url"]; ok {
				proxy := reversproxy.NewReversProxyForTargetUrl(getTargetUrlFromUrlString(urlQueryParam[0]))
				proxy.Director = reversproxy.NewResourcesDirector(getTargetUrlFromUrlString(urlQueryParam[0]), resPath)
				proxy.ServeHTTP(res, req)
			}
		}
	}
}

func xhrProxyHandle(res http.ResponseWriter, req *http.Request) {
	urlFromHeader := req.Header.Get("Referer")
	if targetUrl, err := url.Parse(urlFromHeader); err == nil {
		if urlQueryParam, ok := targetUrl.Query()["url"]; ok {
			proxy := reversproxy.NewReversProxyForTargetUrl(getTargetUrlFromUrlString(urlQueryParam[0]))
			proxy.Director = reversproxy.NewXhrDirector(getTargetUrlFromUrlString(urlQueryParam[0]))
			proxy.ServeHTTP(res, req)
		}
	}
}

func getTargetUrlFromUrlString(urlStr string) *url.URL {
	targetUrl, _ := url.Parse(urlStr)
	return targetUrl
}
