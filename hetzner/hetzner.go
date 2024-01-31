package hetzner

import (
	"context"
	"log"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func DoStuff() {
	token := os.Getenv("HETZNER_API_TOKEN")
	client := hcloud.NewClient(hcloud.WithToken(token))

	server, _, err := client.Server.GetByID(context.Background(), 1)
	if err != nil {
		log.Fatalf("[error-admin] retrieving server: %s", err)
		return
	}
	if server != nil {
		log.Printf("Server 1 is called: %q", server.Name)
	} else {
		log.Printf("Server 1 not found")
	}
}
