package main

import (
	"crypto"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"time"
)

// Returns a custom timestamp (cts) for time.Now() as Day/HH:MM:SS
func cts() string {
	t := time.Now()
	return fmt.Sprintf("%02d/%02d%02d%02d", t.Day(), t.Hour(), t.Minute(), t.Second())
}

// Decodes and unmarshals the JSON response body into the provided destination,
// or fatally exit upon error.
func unmarshalOrExit(body io.Reader, dst interface{}) {
	// User a decoder to read from the response body, and unmarshal the JSON.
	decoder := json.NewDecoder(body)
	// Ensure JSON doesn't contain unexpected fields (not present responseBody).
	decoder.DisallowUnknownFields()
	err := decoder.Decode(dst)
	if err != nil {
		fmt.Printf("[err][admin] decoding & unmarshaling JSON: %v [%s]\n", err, cts())
		os.Exit(1)
	}
}

// Reads from PEM file and sets the global private key variable.
func setPrivateKey() {
	// If the private key file does not exist, exit the program.
	_, err := os.Stat("cp.pem")
	if os.IsNotExist(err) {
		fmt.Printf("[err][admin] private key file is not present; use Provision Local menu to generate and copy to api server: %v [%s]\n", err, cts())
		return
	} else {
		// The private key file exists; read it and set global variable.
		var privateKeyPEM []byte
		privateKeyPEM, err := os.ReadFile("cp.pem")
		if err != nil {
			fmt.Printf("[err][admin] reading private key file: %v [%s]\n", err, cts())
			os.Exit(1)
		}

		// Decode the PEM file into a private key.
		block, _ := pem.Decode(privateKeyPEM)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			fmt.Printf("[err][admin] decoding PEM block containing private key [%s]\n", cts())
			os.Exit(1)
		}

		cpPrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			fmt.Printf("[err][admin] parsing encoded private key: %v [%s]\n", err, cts())
			os.Exit(1)
		}
	}
}

// Returns a base64Url encoded signature of the message.
func signMessage(msg string) string {
	// Make sure private key is present in-memory (global variable).
	if cpPrivateKey == nil {
		fmt.Printf("[err][admin] global private key variable has not been set [%s]\n", cts())
		os.Exit(1)
	}

	// Compute hash of the message.
	hash := sha256.New()
	hash.Write([]byte(msg))
	hashedMessage := hash.Sum(nil)

	// Sign the hashed message.
	signature, err := rsa.SignPKCS1v15(cryptoRand.Reader, cpPrivateKey, crypto.SHA256, hashedMessage)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(signature)
}

// Sets the adminAuthToken global variable.
func setAdminAuthToken() {
	// Make sure admin one's ULID is present in the environment.
	adminUlid := os.Getenv("ADMIN_ONE_ULID")
	if adminUlid == "" {
		fmt.Printf("[err][admin] env variable ADMIN_ONE_ULID is not set [%s]\n", cts())
		os.Exit(1)
	}

	signedAdminUlid := signMessage(adminUlid)
	adminAuthToken = fmt.Sprintf("%s.%s", adminUlid, signedAdminUlid)
}
