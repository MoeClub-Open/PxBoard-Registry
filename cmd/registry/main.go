package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8091", "监听地址")
	root := flag.String("root", ".", "插件源目录")
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /index.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(*root, "index.json"))
	})
	mux.HandleFunc("GET /packages/{file}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		if name != filepath.Base(name) {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(*root, "packages", name))
	})
	log.Printf("registry: http://%s/index.json", *listen)
	s := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(s.ListenAndServe())
}
