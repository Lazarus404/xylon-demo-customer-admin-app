package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/config"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/fleetpoll"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/historystore"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/httpapi"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/interfaceclient"
	"github.com/Lazarus404/xylon-demo-customer-admin-app/internal/webhookstore"
)

func main() {
	configPath := flag.String("config", "config.example.json", "path to JSON config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	hist, err := historystore.Open(cfg.SQLitePath)
	if err != nil {
		log.Fatal(err)
	}
	defer hist.Close()

	iface := interfaceclient.New(cfg.InterfaceBaseURL, cfg.CustomerID)
	recs := webhookstore.New(16)
	poller := &fleetpoll.Poller{
		Interface:            iface,
		History:              hist,
		Recommendations:      recs,
		ScaleOutParticipants: cfg.ScaleOutParticipants,
		Interval:             time.Duration(cfg.FleetPollSec) * time.Second,
	}
	poller.Start()
	defer poller.Stop()

	srv := &httpapi.Server{
		Cfg:             cfg,
		Interface:       iface,
		Webhooks:        webhookstore.New(32),
		Recommendations: recs,
		History:         hist,
	}
	log.Printf("listening on http://%s (customer_id=%s, interface=%s, sqlite=%s, fleet_poll=%ds, scale_out=%d)",
		cfg.ListenAddr, cfg.CustomerID, cfg.InterfaceBaseURL, cfg.SQLitePath, cfg.FleetPollSec, cfg.ScaleOutParticipants)
	if cfg.WebhookSecret != "" {
		log.Printf("webhook ingest enabled (POST /v1/webhooks/xylon)")
	}
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
