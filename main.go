package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Opt struct {
	Port           int
	StaticDir      string
	StaticBasePath string
	ApiTarget      string
	ApiBasePath    string
	Verbose        bool
}

var opt Opt

func init() {
	flag.IntVar(&opt.Port, "p", 9090, "server listening port")
	flag.StringVar(&opt.StaticDir, "s", "", "directory of static files to proxy")
	flag.StringVar(&opt.StaticBasePath, "sp", "/web", "base path of static files")
	flag.StringVar(&opt.ApiTarget, "t", "", "target of api to proxy, eg: http://127.0.0.1:8080")
	flag.StringVar(&opt.ApiBasePath, "tp", "/api", "base path of api to proxy")
	flag.BoolVar(&opt.Verbose, "v", false, "log verbose")
}

func main() {
	flag.Parse()
	if opt.ApiBasePath == opt.StaticBasePath {
		panic("api base path \"tp\" should not be the same as static file base path \"sp\"")
	}

	// api proxy
	if opt.ApiTarget != "" {
		targetUrl, err := url.Parse(opt.ApiTarget)
		if err != nil {
			panic(err)
		}
		proxyApi(targetUrl)
	}

	// static file server
	if opt.StaticDir != "" {
		log.Printf("host static dir %s , base path %s", opt.StaticDir, opt.StaticBasePath)
		f := http.FileServer(http.Dir(opt.StaticDir))
		http.Handle(opt.StaticBasePath,
			http.StripPrefix(opt.StaticBasePath, f))
	}

	log.Printf("Starting Listening on %d", opt.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", opt.Port), nil))
}

func proxyApi(targetUrl *url.URL) {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", targetUrl.Host)
			req.Host = targetUrl.Host
			req.URL.Scheme = targetUrl.Scheme
			req.URL.Host = targetUrl.Host
			if opt.ApiBasePath != "/" {
				req.URL.Path = opt.ApiBasePath + req.URL.Path
			}
		},
		ErrorLog: log.Default(),
	}
	log.Printf("proxy api, target %s base path: %s ", opt.ApiTarget, opt.ApiBasePath)
	http.HandleFunc(opt.ApiBasePath, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request %s %s\n", r.Method, r.URL)
		proxy.ServeHTTP(w, r)
	})

}
