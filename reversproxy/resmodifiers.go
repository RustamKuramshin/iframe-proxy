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
			//doc.Find("head").PrependHtml(
			//	"<script>!function(i){XMLHttpRequest.prototype.open=function(t,e,o,n,s){this.addEventListener(\"readystatechange\",function(){4==this.readyState&&console.log(this.status)},!1),i.call(this,t,e,o,n,s),this.setRequestHeader(\"X-Mark\",\"to-root\")}}(XMLHttpRequest.prototype.open);</script>")
			doc.Find("head").PrependHtml(
				"<script>!function(n){XMLHttpRequest.prototype.open=function(t,e,o,p,s){n.call(this,t,\"/xhrproxy\"+e,o,p,s),this.setRequestHeader(\"X-Mark\",\"to-root\")}}(XMLHttpRequest.prototype.open);</script>")
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
