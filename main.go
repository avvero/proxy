package main

import (
	_ "net/http/pprof"
	"net/http"
	"log"
	"flag"
	"net/url"
	"net/http/httputil"
)

var (
	httpPort = flag.String("httpPort", "8080", "http server port")
)

func main() {
	flag.Parse()

	// proxy stuff
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/proxy", func(response http.ResponseWriter, request *http.Request) {

		target, ok := request.URL.Query()["target"]
		if !ok || len(target) < 1 {
			http.Error(response, "Url Param 'target' is missing", http.StatusInternalServerError)
			return
		}
		url, err := url.Parse(target[0])
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Proxing: " + target[0])

		// create the reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(url)

		// Update the headers to allow for SSL redirection
		request.URL.Host = url.Host
		request.URL.Scheme = url.Scheme
		request.URL.Path = ""
		request.URL.RawQuery = ""
		request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
		request.Host = url.Host
		request.RequestURI = "/"

		response.Header().Set("Access-Control-Allow-Origin", "*")
		response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		response.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Note that ServeHttp is non blocking and uses a go routine under the hood
		proxy.ServeHTTP(response, request)
	})

	log.Println("Http server started on port " + *httpPort)
	http.ListenAndServe(":" + *httpPort, nil)
}
