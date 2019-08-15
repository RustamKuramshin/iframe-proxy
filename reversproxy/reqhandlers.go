package reversproxy

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// IframeProxyHandle обработчик запроса на получение документа для iframe
func IframeProxyHandle(res http.ResponseWriter, req *http.Request) {

	proxy := NewReversProxy()

	// Проксирование EventSource, который используется для стриминга в админ панели
	if req.Header["Accept"][0] == "text/event-stream" {

		// UDF streaming
		targetUrl := ""
		resource := strings.Replace(req.URL.Path, "/iframeproxy/", "", 1)
		if r, err := regexp.Compile("/iframeproxy/([[:alnum:]]{32})/"); err == nil {
			urlhash := r.FindStringSubmatch(req.Header["Referer"][0])
			targetUrl = AdminPanelsUrlHashes[urlhash[1]]
		}
		resource = transformResourceForUdfProxies(targetUrl, resource)
		resource = "/" + resource

		proxy.Director = NewResourceDirector(getTargetUrlFromUrlString(targetUrl), resource)

	} else {

		// Проксирование получения документа для iframe и связанных ресурсов
		targetUrl := AdminPanelsUrlHashes[mux.Vars(req)["urlhash"]]
		resource := mux.Vars(req)["resource"]

		proxy.Director = NewPlusPathDirector(getTargetUrlFromUrlString(targetUrl), resource)

		if resource == "" {
			// Запрос документа для iframe
			proxy.ModifyResponse = NewModifyResponseChangeXhrAndWSConnBehavior(mux.Vars(req)["urlhash"], targetUrl)
		}
	}

	proxy.ServeHTTP(res, req)
}

// XhrProxyHandle обработчик xhr-запросов из iframe
func XhrProxyHandle(res http.ResponseWriter, req *http.Request) {

	targetUrl := AdminPanelsUrlHashes[mux.Vars(req)["urlhash"]]
	resource := mux.Vars(req)["resource"]
	resource = transformResourceForUdfProxies(targetUrl, resource)
	resource = "/" + resource

	proxy := NewReversProxy()
	proxy.Director = NewResourceDirector(getTargetUrlFromUrlString(targetUrl), resource)

	proxy.ServeHTTP(res, req)
}

// WSProxyHandle обработчик проксирования соеденения websocket из iframe
func WSProxyHandle(res http.ResponseWriter, req *http.Request) {

	targetUrl := AdminPanelsUrlHashes[mux.Vars(req)["urlhash"]]
	wsUrlOrigStr := req.URL.Query()["wsurl"][0]
	wsUrlOrig, _ := url.Parse(wsUrlOrigStr)

	proxy := NewReversProxy()
	proxy.Director = NewResourceDirector(getTargetUrlFromUrlString(targetUrl), wsUrlOrig.Path)
	proxy.ServeHTTP(res, req)
}

// getTargetUrlFromUrlString преобразует url-строку в объект *url.URL
func getTargetUrlFromUrlString(urlStr string) (targetUrl *url.URL) {
	targetUrl, _ = url.Parse(urlStr)
	return
}

// transformResourceForUdfProxies преобразует url path для проксирования админ панелей UDF local proxies
// и UDF staging. Например, url, начинающийся с "/admin/ ... " преобразуется в "/udf-proxy-1/admin/ ..."
func transformResourceForUdfProxies(targetUrl string, resource string) string {

	// Для UDF local proxies
	if strings.Contains(targetUrl, "udf-proxy-1") {
		resource = "udf-proxy-1/" + resource
	} else if strings.Contains(targetUrl, "udf-proxy-2") {
		resource = "udf-proxy-2/" + resource
	}

	// Для UDF staging
	if strings.Contains(targetUrl, "hub01-r") {
		resource = "hub01-r/" + resource
	} else if strings.Contains(targetUrl, "hub01-rs") {
		resource = "hub01-rs/" + resource
	} else if strings.Contains(targetUrl, "hub01-rt") {
		resource = "hub01-rt/" + resource
	} else if strings.Contains(targetUrl, "hub01-p") {
		resource = "hub01-p/" + resource
	}

	return resource
}
