package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

// Returns error if the API server is already running.
func apiServerOffline() error {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err == nil {
		conn.Close()
		return fmt.Errorf("api server is already running")
	}
	return nil
}

// Creates temporary directory for the end-to-end test. Prompts user to delete
// if the directory already exists.
func prepareDirectory(dir string) error {
	// Check if the temp directory exists.
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		// Prompt the user.
		fmt.Print("Directory already exists. Do you want to delete it? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		// Reads until the first occurrence of newline delimiter.
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[err][admin] reading user input: %v [%s]\n", err, cts())
			os.Exit(1)
		}

		// If yes, delete the directory.
		if input == "y\n" || input == "Y\n" {
			err = os.RemoveAll(dir)
			if err != nil {
				fmt.Printf("[err][admin] deleting existing directory: %v [%s]\n", err, cts())
				os.Exit(1)
			}
			fmt.Printf("[admin] directory deleted [%s]\n", cts())
		} else {
			// If no, return error.
			return fmt.Errorf("user declined to delete existing directory")
		}
	}

	// Create a new temp directory.
	err := os.Mkdir(dir, 0755)
	if err != nil {
		fmt.Printf("[err][admin] creating temp directory: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	fmt.Printf("[admin] new directory created: %v [%s]\n", dir, cts())

	return nil
}

func runEndToEndLocal() {
	// The API server will be started in a subprocess below. If it is already
	// running in another process, abort this test.
	err := apiServerOffline()
	if err != nil {
		fmt.Printf("[err][admin] confirming server is offline: %v [%s]\n", err, cts())
		return
	}

	// Prepare a temp directory for the test.
	dir := "temp-e2e"
	err = prepareDirectory(dir)
	if err != nil {
		fmt.Printf("[err][admin] preparing directory: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Git clone API into the temp directory.
	goGetCmd := exec.Command("git", "clone", "-q", "https://github.com/reddhouse/cp-api")
	goGetCmd.Dir = dir
	goGetCmd.Stdout = os.Stdout
	goGetCmd.Stderr = os.Stderr

	// Run command and wait for it to complete.
	err = goGetCmd.Run()
	if err != nil {
		fmt.Printf("[err][admin] running git clone command (silently): %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Setup command to start the cp-api server in a subprocess.
	subDir := "cp-api"
	runCmd := exec.Command("go", "run", ".", "-env=dev")
	runCmd.Dir = fmt.Sprintf("%s/%s", dir, subDir)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	// Start server but don't wait in order to proceed with testing.
	err = runCmd.Start()
	if err != nil {
		fmt.Printf("[err][admin] starting an exec.Command: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	fmt.Printf("[admin] subprocess exec.Command has PID: %d [%s]\n", runCmd.Process.Pid, cts())

	// Delay a bit while server starts.
	for i := 0; i < 10; i++ {
		err := apiServerOffline()
		if err != nil {
			break
		}
		// If the connection failed, wait for 1 second before trying again.
		time.Sleep(1 * time.Second)
	}

	// Proceed with testing endpoints.
	signup(generateRandomEmailAddress())
	shutdown()

	// Wait for previously started command to exit.
	err = runCmd.Wait()
	if err != nil {
		fmt.Printf("[err][admin] waiting for exec.Command to exit: %v [%s]\n", err, cts())
		os.Exit(1)
	}
}
