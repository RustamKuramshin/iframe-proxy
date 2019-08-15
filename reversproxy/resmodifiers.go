package reversproxy

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

// NewModifyResponseOverwriteRelPaths - modifier ответа реверс-прокси.
// Изменяет тело http-ответа (html документ).
// Дописывает в src и href атрибуты тегов script и link пути для проксирования запросов ресурсов.
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
		SetHTTPResponseBody(res, html)
		return nil
	}

	return modifyResponseOverwriteRelPaths
}

// NewModifyResponseChangeXhrAndWSConnBehavior - modifier ответа реверс-прокси.
// Изменяет тело http-ответа (html документ).
// Инжектирует в html js-скрипты, которые изменяют базовое поведение XHR-зарпосов и
// подключение websocket.
func NewModifyResponseChangeXhrAndWSConnBehavior(urlhash string, targetUrl string) func(r *http.Response) error {

	var modifyResponseChangeXhrBehavior = func(res *http.Response) error {

		switch res.StatusCode {
		case http.StatusNotFound, http.StatusBadGateway:
			SetHTTPResponseBody(res, fmt.Sprintf("Не удалось загрузить страницу по адресу: %v", targetUrl))
			return nil
		}

		res.Header.Del("X-Frame-Options")

		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)

		if err == nil {

			jsInjectWSandXhrInterceptor := fmt.Sprintf(`<script>
    // Для решения существующей проблемы с относительными путями в страницах админ-панелей,
    // которые загружаются в iframe, был придуман механизм инжектирования этого скрипта на
    // проксе-сервере в документ возварщаемый в iframe.
    
    //WebSocket interceptor
    var _WS = WebSocket;
    WebSocket = function(url, protocols) {
        let urlHash = "%v";
        let WSObject;
        
        let wsProtocol = window.location.protocol == "https:" ? "wss" : "ws";
        let newUrl = wsProtocol + "://" + location.host + "/wsproxy/" + urlHash + "/?wsurl=" + url;
        
        this.url = newUrl;
        this.protocols = protocols;
        if (!this.protocols) { 
            WSObject = new _WS(newUrl) 
        } else { 
            WSObject = new _WS(newUrl, protocols)
        }
        return WSObject;
    };

   //XHR interceptor
   (function (open) {
       XMLHttpRequest.prototype.open = function (method, url, async, user, pass) {

            //console.log("PROXY DEBUG :: url BEFORE - " + url);
           
            let urlHash = "%v";
           
            // UDF BEGIN
            let newUrl = url.replace("..", "");
            
            if (/^pages/i.test(newUrl) || /^templates/i.test(newUrl)){
               newUrl = "/admin/" + newUrl;
            }
            // UDF END
            
            // ALL
            newUrl = "/xhrproxy/" + urlHash + newUrl;
            // ALL

            //console.log("PROXY DEBUG :: url AFTER - " + newUrl);

           open.call(this, method, newUrl, async, user, pass);
           this.setRequestHeader("X-Target-Url", urlHash)
       };
   })(XMLHttpRequest.prototype.open);</script>`, urlhash, urlhash)

			doc.Find("head").PrependHtml(jsInjectWSandXhrInterceptor)

			jsInjectUrlPar := fmt.Sprintf(`<script>
           let body = document.querySelector("body");
           let par = document.createElement("p");
           par.textContent = "%v";
           body.insertBefore(par, body.firstChild);</script>`, targetUrl)

			doc.Find("body").PrependHtml(jsInjectUrlPar)
		}

		html, _ := doc.Html()

		SetHTTPResponseBody(res, html)

		return nil
	}

	return modifyResponseChangeXhrBehavior
}

// NewModifyResponseCutXFrame - modifier ответа реверс-прокси.
// Изменяет тело http-ответа (html документ).
// Просто вырезает заголовок, запрещающий открытие документа в iframe.
func NewModifyResponseCutXFrame() func(r *http.Response) error {

	var cutXFrame = func(res *http.Response) error {
		res.Header.Del("X-Frame-Options")
		return nil
	}
	return cutXFrame
}
