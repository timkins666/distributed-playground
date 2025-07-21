package main

import (
	"log"
	"net/http"
	"os"
)

var proxyHosts map[string]string = map[string]string{
	"auth":    os.Getenv("AUTH_SERVICE_HOST"),
	"account": os.Getenv("ACCOUNT_SERVICE_HOST"),
	"payment": os.Getenv("PAYMENT_SERVICE_HOST"),
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)

	for srv := range proxyHosts {
		mux.HandleFunc("/"+srv+"/", withAuth(proxyToService))
	}

	port := ":" + os.Getenv("SERVE_PORT")
	log.Println("API Gateway running on", port)
	log.Fatal(http.ListenAndServe(port, corsMiddleware(mux)))
}
