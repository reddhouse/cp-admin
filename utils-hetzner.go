package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"gopkg.in/yaml.v3"
)

type UserData struct {
	Users          []User   `yaml:"users"`
	PackageUpdate  bool     `yaml:"package_update"`
	PackageUpgrade bool     `yaml:"package_upgrade"`
	Packages       []string `yaml:"packages"`
	WriteFiles     []File   `yaml:"write_files"`
	RunCmd         []string `yaml:"runcmd"`
}

type User struct {
	Name              string   `yaml:"name"`
	Groups            string   `yaml:"groups"`
	Sudo              string   `yaml:"sudo"`
	Shell             string   `yaml:"shell"`
	LockPasswd        bool     `yaml:"lock_passwd"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

type File struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Owner       string `yaml:"owner"`
	Permissions string `yaml:"permissions"`
	Defer       bool   `yaml:"defer"`
}

func hetznerViewCurrentResources() {
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

	// Print all servers.
	for _, server := range servers {
		fmt.Printf("[admin] server ID: %d, name: %s, status: %s [%s]\n", server.ID, server.Name, server.Status, cts())
	}
}

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
		PackageUpdate:  true,
		PackageUpgrade: true,
		Packages: []string{
			"nginx",
			"snapd",
			"fail2ban",
			"ufw",
			"unzip",
		},
		WriteFiles: []File{
			{
				Path: "/etc/nginx/sites-available/cp",
				Content: `server {
    server_name cooperativeparty.org www.cooperativeparty.org;
    listen 80;
    location /api {
        proxy_set_header   X-Forwarded-For $remote_addr;
        proxy_set_header   Host $http_host;
        proxy_pass         "http://127.0.0.1:8000";
    }
}`,
				// An empty string sets owner to default (root).
				// Owner: nginx:nginx was causing an error in cloud-init,
				// perhaps due to nginx package installation taking too long
				// before the write_files module was run.
				Owner: "",
				// When Owner: nginx:nginx, I had permissions set to "0640".
				// Now I want to make sure nginx process can read the file, even
				// if it's not the owner, so set to 0644.
				Permissions: "0644",
				Defer:       true,
			},
		},
		RunCmd: []string{
			"systemctl enable nginx",
			"ufw allow 'Nginx Full'",
			`printf "[sshd]\nenabled = true\nbanaction = iptables-multiport" > /etc/fail2ban/jail.local`,
			"systemctl enable fail2ban",
			"systemctl start fail2ban",
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

func hetznerCreateServer1() {
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
	fmt.Printf("[admin] created server with ID: %v [%s]\n", result.Server.ID, cts())
}
