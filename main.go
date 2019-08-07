package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"

	"proxyAdminPanels/reversproxy"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/pageproxy", pageProxyHandle).Methods("GET")
	router.HandleFunc("/resourceproxy", transparentProxyHandle).Methods("GET")

	router.HandleFunc("/proxy", transparentProxyHandle).Methods("GET")

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	log.Println("Server starting")
	// log.Fatal(http.ListenAndServeTLS(":9090", "cert.pem", "key.pem", router))
	log.Fatal(http.ListenAndServe(":9090", router))
}

func pageProxyHandle(res http.ResponseWriter, req *http.Request) {
	targetUrlStr := req.URL.Query()["url"][0]
	proxy := reversproxy.NewReversProxy(getTargetUrlFromUrlString(targetUrlStr), reversproxy.NewModifyResponseOverwriteRelPaths(targetUrlStr))
	proxy.ServeHTTP(res, req)
}

func transparentProxyHandle(res http.ResponseWriter, req *http.Request) {
	proxy := reversproxy.NewReversProxy(getTargetUrlFromUrlString(req.URL.Query()["url"][0]), reversproxy.NewModifyResponseCutXFrame())
	proxy.ServeHTTP(res, req)
}

func getTargetUrlFromUrlString(urlStr string) *url.URL {
	targetUrl, _ := url.Parse(urlStr)
	return targetUrl
}
