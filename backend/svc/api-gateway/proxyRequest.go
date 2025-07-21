package main

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func proxyToService(w http.ResponseWriter, r *http.Request) {
	log.Println("request path: ", r.URL.Path)

	service, path, found := strings.Cut(r.URL.Path[1:], "/")
	if !found {
		http.Error(w, "Unknown", http.StatusInternalServerError)
	}

	log.Println("request service:", service)

	host, ok := proxyHosts[strings.ToLower(service)]
	if !ok {
		log.Printf("Requested unknown service %s", service)
		http.Error(w, "Unknown", http.StatusNotFound)
	}

	req, err := http.NewRequest(r.Method, host+"/"+path, r.Body)
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers (like Authorization) to the proxied request
	req.Header = r.Header.Clone()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, "Failed to contact account service", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
