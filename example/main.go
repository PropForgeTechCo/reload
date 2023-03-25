package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/aarol/reload"
)

func newTemplateCache() *template.Template {
	return template.Must(template.ParseGlob("ui/*.html"))
}

func main() {
	templateCache := newTemplateCache()

	reload.OnReload = func() {
		templateCache = newTemplateCache()
	}

	// serve any static files like you normally would
	http.Handle("/static/", http.FileServer(http.Dir("ui/")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// serve index.html with template data
		data := map[string]any{
			"Timestamp": time.Now().Format(time.RFC850),
		}
		err := templateCache.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			fmt.Println(err)
		}
	})
	// isDevelopment := os.Getenv("MODE") == "development"
	isDevelopment := true

	// this can be any http.Handler like mux.Router or chi.Router
	var handler http.Handler = http.DefaultServeMux

	if isDevelopment {
		handler = reload.WatchAndInject("ui/")(handler)
	}

	addr := "localhost:3001"

	fmt.Println("Server running at", addr)

	http.ListenAndServe(addr, handler)
}
