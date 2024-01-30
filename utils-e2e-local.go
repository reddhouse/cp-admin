package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

func runEndToEndLocal() {
	dir := "temp-e2e"

	// Check if the temp directory exists.
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		// Prompt the user
		fmt.Print("[admin] Directory already exists. Do you want to delete it? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		// Reads until the first occurrence of newline delimiter.
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// If yes, delete the directory.
		if input == "y\n" || input == "Y\n" {
			err = os.RemoveAll(dir)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("[admin] Directory deleted.")
		} else {
			// If no, exit.
			return
		}
	}

	// Create a new temp directory.
	err := os.Mkdir(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[admin] New %v directory created.\n", dir)

	// Git clone API into the temp directory.
	goGetCmd := exec.Command("git", "clone", "https://github.com/reddhouse/cp-api")
	goGetCmd.Dir = dir
	goGetCmd.Stdout = os.Stdout
	goGetCmd.Stderr = os.Stderr

	// Run command and wait for it to complete.
	err = goGetCmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Setup command to start the cp-api server in a subprocess.
	subDir := "cp-api"
	runCmd := exec.Command("/Users/jmt/sdk/go1.22rc1/bin/go", "run", ".", "-test")
	runCmd.Dir = fmt.Sprintf("%s/%s", dir, subDir)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	// Start server but don't wait in order to proceed with testing.
	err = runCmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("[admin] Subprocess exec.Command (cp-admin) has PID: %d\n", runCmd.Process.Pid)

	// Check if the server is up by trying to establish a connection to it.
	for i := 0; i < 10; i++ {
		conn, err := net.Dial("tcp", "localhost:8001")
		if err == nil {
			conn.Close()
			break
		}
		// If the connection failed, wait for 1 second before trying again.
		time.Sleep(1 * time.Second)
	}

	// Set global port variable.
	port = "8001"

	// Proceed with testing endpoints.
	signup()
	shutdown()

	// Reset global port variable.
	port = "8000"

	// Wait for previously started command to exit.
	err = runCmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
