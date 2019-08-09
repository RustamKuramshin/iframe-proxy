package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"proxyAdminPanels/reversproxy"
)

func main() {
	router := mux.NewRouter()

	//router.HandleFunc("/pageproxy", pageProxyHandle).Methods("GET")
	//router.HandleFunc("/resourceproxy", transparentProxyHandle).Methods("GET")

	//router.PathPrefix("/").Headers("X-Requested-With", "XMLHttpRequest").HandlerFunc(xhrProxyHandle)
	//router.PathPrefix("/").Headers("Accept", "application/json, text/plain, */*").HandlerFunc(xhrProxyHandle)
	//router.PathPrefix("/").Headers("Accept", "text/event-stream").HandlerFunc(xhrProxyHandle)

	//router.PathPrefix("/").Headers("X-Mark", "to-root").HandlerFunc(xhrProxyHandle)

	router.PathPrefix("/iframeproxy/").Methods("GET").HandlerFunc(reversproxy.IframeProxyHandle)
	router.PathPrefix("/xhrproxy").Methods("GET").HandlerFunc(reversproxy.XhrProxyHandle)
	router.PathPrefix("/transparentxhrproxy").Methods("GET").HandlerFunc(reversproxy.TransparentXhrProxyHandle)

	//router.HandleFunc("/proxy", transparentProxyHandle).Methods("GET")

	//router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(http.Dir("./web/"))))

	log.Println("Server starting")
	// log.Fatal(http.ListenAndServeTLS(":9090", "cert.pem", "key.pem", router))
	log.Fatal(http.ListenAndServe(":9090", router))
}
