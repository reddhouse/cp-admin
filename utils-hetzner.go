package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"gopkg.in/yaml.v3"
)

type UserData struct {
	Users          []User    `yaml:"users"`
	AptUpgrade     bool      `yaml:"apt_upgrade"`
	Apt            AptConfig `yaml:"apt"`
	PackageUpdate  bool      `yaml:"package_update"`
	PackageUpgrade bool      `yaml:"package_upgrade"`
	Packages       []string  `yaml:"packages"`
	WriteFiles     []File    `yaml:"write_files"`
	RunCmd         []string  `yaml:"runcmd"`
}

type User struct {
	Name              string   `yaml:"name"`
	Groups            string   `yaml:"groups"`
	Sudo              string   `yaml:"sudo"`
	Shell             string   `yaml:"shell"`
	LockPasswd        bool     `yaml:"lock_passwd"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

type AptConfig struct {
	Sources map[string]SourceConfig `yaml:"sources"`
}

type SourceConfig struct {
	Source string `yaml:"source"`
}

type File struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Owner       string `yaml:"owner"`
	Permissions string `yaml:"permissions"`
	Defer       bool   `yaml:"defer"`
}

// Global variable to hold server names and hcloud server instances
var serverMap map[string]*hcloud.Server = make(map[string]*hcloud.Server)

// Uses the Hetzner API client to grab known resources and store server info in a local variable.
func hetznerGetAndSetCurrentResources() {
	// SSH Key(s)
	sshKeys, err := hcloudClient.SSHKey.All(context.TODO())
	if err != nil {
		fmt.Printf("[err][admin] retrieving ssh key(s): %s [%s]\n", err, cts())
		os.Exit(1)
		return
	}

	// If sshKeys is empty, print message and return.
	if len(sshKeys) == 0 {
		fmt.Printf("[admin] no servers found [%s]\n", cts())
		return
	}

	// Print all servers.
	for _, key := range sshKeys {
		fmt.Printf("[admin] ssh key ID: %d, name: %s [%s]\n", key.ID, key.Name, cts())
	}

	// Servers
	servers, err := hcloudClient.Server.All(context.TODO())
	if err != nil {
		fmt.Printf("[err][admin] retrieving servers: %s [%s]\n", err, cts())
		os.Exit(1)
		return
	}

	// If servers is empty, print message and return.
	if len(servers) == 0 {
		fmt.Printf("[admin] no servers found [%s]\n", cts())
		return
	}

	// Set servers in global serverMap variable (name:server mapping).
	for _, server := range servers {
		serverMap[server.Name] = server
	}

	// Print all servers.
	for _, server := range serverMap {
		fmt.Printf("[admin] server ID: %d, ip: %s, name: %s, status: %s [%s]\n", server.ID, server.PublicNet.IPv4.IP, server.Name, server.Status, cts())
	}
}

// Creates a new SSH key on Hetzner cloud.
func hetznerCreateSSHKey() {
	pubKeyPath := os.Getenv("LOCAL_PUBLIC_KEY_PATH")
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		fmt.Printf("[err][admin] reading local public key file at: %s: %v [%s]\n", pubKeyPath, err, cts())
		os.Exit(1)
	}

	// Define SSH key options.
	opts := hcloud.SSHKeyCreateOpts{
		Name:      os.Getenv("HETZNER_PUBLIC_KEY_NAME"),
		PublicKey: string(pubKey),
	}

	// Create SSH key.
	sshKey, _, err := hcloudClient.SSHKey.Create(context.TODO(), opts)
	if err != nil {
		fmt.Printf("[err][admin] creating SSH key: %v [%s]\n", err, cts())
		return
	}

	// Print the ID of the created SSH key.
	fmt.Printf("[admin] created SSH key with ID: %v [%s]", sshKey.ID, cts())
}

// Creates a yaml formatted string of "user data" for cloud-init.
func createUserData() string {
	pubKeyPath := os.Getenv("LOCAL_PUBLIC_KEY_PATH")
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		fmt.Printf("[err][admin] reading local public key file at: %s: %v [%s]\n", pubKeyPath, err, cts())
		os.Exit(1)
	}

	userData := UserData{
		Users: []User{
			{
				Name:   os.Getenv("CP_ADMIN_USER_ONE"),
				Groups: "users, admin",
				Sudo:   "ALL=(ALL) NOPASSWD:ALL",
				Shell:  "/bin/bash",
				// Prevents user from logging in using password authentication.
				LockPasswd:        true,
				SshAuthorizedKeys: []string{string(pubKey)},
			},
		},
		AptUpgrade: true,
		Apt: AptConfig{
			Sources: map[string]SourceConfig{
				"caddy": {
					Source: "deb [trusted=yes] https://dl.cloudsmith.io/public/caddy/stable/deb/ubuntu jammy main",
				},
			},
		},
		PackageUpdate:  true,
		PackageUpgrade: true,
		Packages: []string{
			"caddy",
			"ufw",
			"unzip",
		},
		WriteFiles: []File{
			{
				Path: "/etc/caddy/Caddyfile",
				Content: fmt.Sprintf(`www.cooperativeparty.org {
	redir https://cooperativeparty.org{uri} permanent
}

cooperativeparty.org {
	tls %s
	header Content-Type text/html
	respond <<HTML
		<html>
			<head><title>Foo</title></head>
			<body>Foo</body>
		</html>
		HTML 200

}`, os.Getenv("CP_ADMIN_USER_ONE_EMAIL")),
				// An empty string sets owner to default (root).
				Owner: "",
				// Allow owner to read and write, and the group/others to read.
				Permissions: "0644",
				// No reason to wait until final stage of cloud-init to write.
				Defer: false,
			},
		},
		RunCmd: []string{
			// Enable Caddy to start on boot and start it immediately (now flag).
			"systemctl enable --now caddy",
			// Allow incoming traffic on HTTP (80), HTTPS (443), and SSH (22) ports.
			"ufw allow http",
			"ufw allow https",
			"ufw allow 'OpenSSH'",
			"ufw enable",
			// Disallow root login.
			"sed -ie '/^PermitRootLogin/s/^.*$/PermitRootLogin no/' /etc/ssh/sshd_config",
			// Disallow password authentication.
			"sed -ie '/^#PasswordAuthentication/s/^.*$/PasswordAuthentication no/' /etc/ssh/sshd_config",
			// Disallow X11 forwarding.
			"sed -ie '/^X11Forwarding/s/^.*$/X11Forwarding no/' /etc/ssh/sshd_config",
			// Disconnect a client after 2 failed authentication attempts.
			"sed -ie '/^#MaxAuthTries/s/^.*$/MaxAuthTries 2/' /etc/ssh/sshd_config",
			// Disallow TCP forwarding.
			"sed -ie '/^#AllowTcpForwarding/s/^.*$/AllowTcpForwarding no/' /etc/ssh/sshd_config",
			// Prevent this (remote) server from using key to authenticate to other servers.
			"sed -ie '/^#AllowAgentForwarding/s/^.*$/AllowAgentForwarding no/' /etc/ssh/sshd_config",
			// Allow only the admin user to SSH into the server.
			fmt.Sprintf("sed -i '$a AllowUsers %s' /etc/ssh/sshd_config", os.Getenv("CP_ADMIN_USER_ONE")),
			// Restart the SSH service to apply changes.
			"systemctl restart ssh",
		},
	}

	data, err := yaml.Marshal(&userData)
	if err != nil {
		fmt.Printf("[err][admin] marshaling userData to yaml: %v [%s]\n", err, cts())
	}
	// Add comment for cloud-init to recognize this file as cloud-config.
	return "#cloud-config\n" + string(data)
}

// Uses createUserData to write a yaml file to disk.
func writeUserDataToFile() {
	userData := createUserData()
	err := os.WriteFile("user_data_test.yml", []byte(userData), 0644)
	if err != nil {
		fmt.Printf("[err][admin] writing user data to file: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	fmt.Printf("[admin] user data successfully written to file [%s]\n", cts())
}

// Create a Hetzner cloud server instance with the name "cp-1".
func hetznerCreateServerOne() {
	// Get the SSH key by name.
	sshKey, _, err := hcloudClient.SSHKey.Get(context.TODO(), os.Getenv("HETZNER_PUBLIC_KEY_NAME"))
	if err != nil {
		fmt.Printf("[err][admin] getting SSH key: %v [%s]\n", err, cts())
		return
	}

	// Define server options.
	opts := hcloud.ServerCreateOpts{
		Name:       "cp-1",
		ServerType: &hcloud.ServerType{Name: "cpx11"},
		Image:      &hcloud.Image{Name: "ubuntu-20.04"},
		Location:   &hcloud.Location{Name: "hil"},
		SSHKeys:    []*hcloud.SSHKey{sshKey},
		UserData:   createUserData(),
	}

	// Create server.
	result, _, err := hcloudClient.Server.Create(context.TODO(), opts)
	if err != nil {
		fmt.Printf("[err][admin] creating server: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Print the ID of the created server
	fmt.Printf("[admin] created server with ID: %v, and IP: %v [%s]\n", result.Server.ID, result.Server.PublicNet.IPv4.IP, cts())
}

// Delete Hetzner cloud server instance that has the name "cp-1".
func hetznerDeleteServerOne() {
	server, ok := serverMap["cp-1"]
	if !ok {
		fmt.Printf("[err][admin] server with name \"cp-1\" not found locally... run Get/Set Current Resources command[%s]\n", cts())
		return
	}

	_, _, err := hcloudClient.Server.DeleteWithResult(context.TODO(), server)
	if err != nil {
		fmt.Printf("[err][admin] deleting server: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	fmt.Printf("[admin] deleted server cp-1 [%s]\n", cts())
	fmt.Printf("[admin] removing known host... [%s]\n", cts())

	// Remove cp-1 IP address from ssh known hosts.
	goGetCmd := exec.Command("ssh-keygen", "-R", server.PublicNet.IPv4.IP.String())
	goGetCmd.Stdout = os.Stdout
	goGetCmd.Stderr = os.Stderr

	// Run command and wait for it to complete.
	err = goGetCmd.Run()
	if err != nil {
		fmt.Printf("[err][admin] running ssh-keygen -R command: %v [%s]\n", err, cts())
		os.Exit(1)
	}
}
