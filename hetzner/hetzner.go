package hetzner

import (
	"context"
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func DoStuff(args ...string) {
	token := os.Getenv("HETZNER_API_TOKEN")
	client := hcloud.NewClient(hcloud.WithToken(token))

	server, _, err := client.Server.GetByID(context.Background(), 1)
	if err != nil {
		fmt.Printf("error retrieving server: %s\n", err)
		return
	}
	if server != nil {
		fmt.Printf("Server 1 is called %q\n", server.Name)
	} else {
		fmt.Printf("Server 1 not found\n")
	}
}
