/*
   Package reversproxy - это реверс-прокси для запросов из тега iframe (issues DASH-130).
   Реализован как обёртка над httputil.ReverseProxy, с поправкой на требования к KeepAlive и пр.
   Данный пакет представлен след. файлами:
   1) reversproxy.go - собственно файл содержащий конструктор структуры httputil.ReverseProxy.
   2) directors.go - набор конструкторов для получения функции-директора, которая должна описывать
   правила преобразования входящего запроса.
   3) reqhandlers.go - набор обработчиков запросов к конечным точкам реверс-прокси.
   4) resmodifiers.go - набор конструкторов фукнций перобразования ответа от реверс-прокси, возвращаемого
   на клиент.
   5) urlhashes.go - инициализирует map для хранения хешей url.

   Маршрут при проксировании запросов:
   A_request => (proxy := NewReversProxy()) => (proxy.Director) => (proxy.ServeHTTP(res, req)) => out_request
   out_response => (proxy.ModifyResponse) => (proxy.ServeHTTP(res, req)) => A_response
*/
package reversproxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

// NewReversProxy базовый конструктор reverse proxy, учитывающий переиспользование соединений, таймауты и keep-alive.
func NewReversProxy() *httputil.ReverseProxy {

	return &httputil.ReverseProxy{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return http.ProxyFromEnvironment(req)
			},
			Dial: func(network, addr string) (net.Conn, error) {
				conn, err := (&net.Dialer{
					Timeout:   60 * time.Second,
					KeepAlive: 60 * time.Second,
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

// PrintResponse выводит в терминал содержимое ответа реверс-прокси,
// который отправляется клиенту. Нужен для отладки.
func PrintResponse(res *http.Response) {
	log.Println("HTTP response")
	if body, err := httputil.DumpResponse(res, true); err == nil {
		fmt.Println(string(body))
	}
}

// SetHTTPResponseBody получает указать на объект http-ответа и
// заменяет его тело на стоку newBody.
// За пределами вызова этой функции нужно сделать "defer res.Body.Close()"
func SetHTTPResponseBody(res *http.Response, newBody string) {
	bodyByte := []byte(newBody)
	res.Body = ioutil.NopCloser(bytes.NewReader(bodyByte))
	res.ContentLength = int64(len(bodyByte))
	res.Header.Set("Content-Length", strconv.Itoa(len(bodyByte)))
}
