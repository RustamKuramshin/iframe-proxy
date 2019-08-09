package reversproxy

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"strconv"
)

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
		PrintResponse(res)
		return nil
	}

	return modifyResponseOverwriteRelPaths

}

func NewModifyResponseChangeXhrBehavior() func(r *http.Response) error {

	var modifyResponseChangeXhrBehavior = func(res *http.Response) error {

		res.Header.Del("X-Frame-Options")

		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)

		if err == nil {

			injectJs := `<script>

    // Для решения существующей проблемы с относительными путями в страницах админ-панелей,
    // которые прогружаются в iframe, был придуман механизм инжектирования этого скрипта на 
    // проксе-сервере в документ возварщаемый в iframe.
    (function (open) {
        XMLHttpRequest.prototype.open = function (method, url, async, user, pass) {
            
            // console.log("PROXY DEBUG :: " + window.location.href);
            // let parseUrl = document.createElement('a');
            // parseUrl.href = window.location.href;
            //
            // const urlParams = new URLSearchParams(parseUrl.search);
            // const targerUrl = urlParams.get('url');
            //
            // let parseTargerUrl = document.createElement('a');
            // parseTargerUrl.href = targerUrl;

            console.log("PROXY DEBUG :: url BEFORE - " + url);

            let newUrl = url.replace("..", "");

            if (newUrl === "pages/dashboard.html") {
                newUrl = "/admin/" + newUrl;
            }
            if (newUrl !== "/stat?groupBy=0") {
                newUrl = "/xhrproxy" + newUrl;
            } else {
                //newUrl = parseTargerUrl.protocol + '//' + parseTargerUrl.hostname + (parseTargerUrl.port ? ':' + parseTargerUrl.port : '') + newUrl;
                newUrl = "/transparentxhrproxy" + newUrl;
            }

            console.log("PROXY DEBUG :: url AFTER - " + newUrl);

            open.call(this, method, newUrl, async, user, pass);
        };
    })(XMLHttpRequest.prototype.open);
    
</script>`

			//doc.Find("head").PrependHtml(
			//	"<script>!function(i){XMLHttpRequest.prototype.open=function(t,e,o,n,s){this.addEventListener(\"readystatechange\",function(){4==this.readyState&&console.log(this.status)},!1),i.call(this,t,e,o,n,s),this.setRequestHeader(\"X-Mark\",\"to-root\")}}(XMLHttpRequest.prototype.open);</script>")
			doc.Find("head").PrependHtml(injectJs)
		}

		html, _ := doc.Html()
		bodyByte := []byte(html)
		res.Body = ioutil.NopCloser(bytes.NewReader(bodyByte))
		res.ContentLength = int64(len(bodyByte))
		res.Header.Set("Content-Length", strconv.Itoa(len(bodyByte)))
		PrintResponse(res)
		return nil
	}

	return modifyResponseChangeXhrBehavior
}

func NewModifyResponseCutXFrame() func(r *http.Response) error {

	var cutXFrame = func(res *http.Response) error {
		res.Header.Del("X-Frame-Options")
		PrintResponse(res)
		return nil
	}

	return cutXFrame
}
