package hetzner

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func DoStuff(args ...string) {
	token := os.Getenv("HETZNER_API_TOKEN")
	client := hcloud.NewClient(hcloud.WithToken(token))

	server, _, err := client.Server.GetByID(context.Background(), 1)
	if err != nil {
		log.Fatalf("error retrieving server: %s\n", err)
	}
	if server != nil {
		fmt.Printf("server 1 is called %q\n", server.Name)
	} else {
		fmt.Println("server 1 not found")
	}
}
