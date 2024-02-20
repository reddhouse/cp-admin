package main

import (
	"bufio"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

func generatePrivateKey() {
	// Check if the private key file exists.
	_, err := os.Stat("cp.pem")
	if os.IsNotExist(err) {
		var err error
		// The private key file does not exist, so generate a new key.
		cpPrivateKey, err := rsa.GenerateKey(cryptoRand.Reader, 2048)
		if err != nil {
			fmt.Printf("[err][admin] creating private key: %v [%s]\n", err, cts())
			os.Exit(1)
		}

		// Encode the private key into PEM format.
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(cpPrivateKey)
		privateKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		})

		// Write the PEM to a file.
		err = os.WriteFile("cp.pem", privateKeyPEM, 0600)
		if err != nil {
			fmt.Printf("[err][admin] writing private key to disk: %v [%s]\n", err, cts())
			os.Exit(1)
		}
		fmt.Printf("[admin] private key successfully created [%s]\n", cts())
	} else {
		// The private key file exists. Ask user to manually delete it.
		fmt.Printf("[err][admin] \"cp.pem\" already exists; delete if you wish to proceed with new key generation [%s]\n", cts())
	}
}

func copyPrivateKeyLocal() {
	// Open the source file for reading
	srcFile, err := os.Open("cp.pem")
	if err != nil {
		fmt.Printf("[err][admin] opening private key file: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(os.Getenv("LOCAL_CP_API_PK_PATH"))
	if err != nil {
		fmt.Printf("[err][admin] creating new file in cp-api directory: %v [%s]\n", err, cts())
	}
	defer dstFile.Close()

	// Use io.Copy to copy the contents of the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Printf("[err][admin] copying old file to new file in cp-api directory: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Call Sync to flush writes to stable storage
	dstFile.Sync()

	fmt.Printf("[admin] private key successfully copied to %s [%s]\n", os.Getenv("LOCAL_CP_API_PK_PATH"), cts())
}

func wrappedCopyPrivateKeyLocal() {
	// Check if the private key file exists in this directory.
	_, err := os.Stat("cp.pem")
	if os.IsNotExist(err) {
		// The private key file does not exist.
		fmt.Printf("[err][admin] \"cp.pem\" does not exist in this directory; generate private key first [%s]\n", cts())
		return
	}

	// Check if a key already exists at destination directory.
	if _, err := os.Stat(os.Getenv("LOCAL_CP_API_PK_PATH")); !os.IsNotExist(err) {
		// Prompt the user.
		fmt.Printf("Local private key already exists at %s. Do you want to overwrite it? (y/n): ", os.Getenv("LOCAL_CP_API_PK_PATH"))
		reader := bufio.NewReader(os.Stdin)
		// Reads until the first occurrence of newline delimiter.
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[err][admin] reading user input: %v [%s]\n", err, cts())
			os.Exit(1)
		}

		// If yes, proceed with key copy.
		if input == "y\n" || input == "Y\n" {
			copyPrivateKeyLocal()
			return
		} else {
			// If no, print message and return.
			fmt.Printf("[admin] user declined to delete existing private key in cp-api directory [%s]\n", cts())
			return
		}
	}
	// No key exists at destination directory so proceed with key copy.
	copyPrivateKeyLocal()
}
