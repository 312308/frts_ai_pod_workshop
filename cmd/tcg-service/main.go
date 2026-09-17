// Command tcg-service is the FRTS Data Ingestion Layer entry point.
//
// SP-01 scope: wire FOP OAuth2 client & token acquisition (BRD FR-024) at
// startup. Entity poll workflows, HTTP handlers, and persistence wiring
// land in later stories (SP-05 onward) per the Approved sprint plan
// frts-ingestion-s1.
package main

import (
	"context"
	"log"

	"tcg-service/internal/config"
	fopclient "tcg-service/internal/fop-client"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("tcg-service: config: %v", err)
	}

	client, err := fopclient.NewClient(context.Background(), cfg)
	if err != nil {
		log.Fatalf("tcg-service: fop client: %v", err)
	}

	log.Printf("tcg-service: FOP client ready (base_url=%s, partner_id=%s)", client.BaseURL, cfg.PartnerID)
}
