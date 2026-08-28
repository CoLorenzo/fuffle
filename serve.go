package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

func serveHTML(htmlContent, resultsHTML, addr string) error {
	if addr == "" {
		addr = ":8080"
	}
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, htmlContent)
	})

	mux.HandleFunc("/results.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, resultsHTML)
	})

	mux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/images/")
		if path == "" || strings.Contains(path, "..") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, path)
	})

 url := "http://localhost" + addr
	fmt.Printf("Serving report at %s\n", url)
	openBrowser(url)

	return http.ListenAndServe(addr, mux)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
