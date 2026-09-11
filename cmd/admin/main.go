package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/config"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/httpapi"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
)

func main() {
	configPath := flag.String("config", "config.example.json", "path to JSON config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	srv := &httpapi.Server{
		Cfg:       cfg,
		Interface: interfaceclient.New(cfg.InterfaceBaseURL, cfg.CustomerID),
	}
	log.Printf("listening on http://%s (customer_id=%s, interface=%s)",
		cfg.ListenAddr, cfg.CustomerID, cfg.InterfaceBaseURL)
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
