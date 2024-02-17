package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func doHetznerStuff() {
	token := os.Getenv("HETZNER_API_TOKEN")
	client := hcloud.NewClient(hcloud.WithToken(token))

	server, _, err := client.Server.GetByID(context.Background(), 1)
	if err != nil {
		fmt.Printf("[err][admin] retrieving server: %s [%s]\n", err, cts())
		os.Exit(1)
		return
	}
	if server != nil {
		fmt.Printf("[admin] server 1 is called: %q [%s]\n", server.Name, cts())
	} else {
		fmt.Printf("[err][admin] server 1 not found [%s]\n", cts())
	}
}
