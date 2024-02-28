package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func doHetznerStuff() {
	server, _, err := hcloudClient.Server.GetByID(context.TODO(), 1)
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

func hetznerCreateSSHKey() {
	pubKeyPath := os.Getenv("LOCAL_PUBLIC_KEY_PATH")
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		fmt.Printf("[err][admin] reading local public key file at: %s: %v [%s]\n", pubKeyPath, err, cts())
		os.Exit(1)
	}

	// Define SSH key options
	opts := hcloud.SSHKeyCreateOpts{
		Name:      os.Getenv("HETZNER_PUBLIC_KEY_NAME"),
		PublicKey: string(pubKey),
	}

	// Create SSH key
	sshKey, _, err := hcloudClient.SSHKey.Create(context.TODO(), opts)
	if err != nil {
		fmt.Printf("[err][admin] creating SSH key: %v [%s]\n", err, cts())
		return
	}

	// Print the ID of the created SSH key
	fmt.Printf("[admin] created SSH key with ID: %v [%s]", sshKey.ID, cts())
}
