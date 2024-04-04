package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Shut down API server gracefully.
func shutdown() {
	url := "http://localhost:8000/api/admin/shutdown/"

	// Create a new request using http.
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("[err][admin] creating request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Set custom admin auth header.
	req.Header.Set("Admin-Authorization", adminAuthToken)

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response body into memory so we can print it.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[err][admin] reading response body: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())
	fmt.Printf("[admin] response body: %s [%s]\n", body, cts())
}

// Log USER_EMAIL bucket on API server.
func logUserEmailBucket() {
	url := "http://localhost:8000/api/admin/log-bucket-custom-key/USER_EMAIL"

	// Create a new request using http.
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("[err][admin] creating request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Set custom admin auth header.
	req.Header.Set("Admin-Authorization", adminAuthToken)

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())
}

// Log USER_AUTH bucket on API server.
func logUserAuthBucket() {
	url := "http://localhost:8000/api/admin/log-bucket/USER_AUTH"
	// Create a new request using http.
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("[err][admin] creating request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Set custom admin auth header.
	req.Header.Set("Admin-Authorization", adminAuthToken)

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())
}

// Log ADMIN_EMAIL bucket on API server.
func logAdminEmailBucket() {
	url := "http://localhost:8000/api/admin/log-bucket/ADMIN_EMAIL"
	// Create a new request using http.
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("[err][admin] creating request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Set custom admin auth header.
	req.Header.Set("Admin-Authorization", adminAuthToken)

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())
}

// Log MOD_EXIM bucket on API server.
func logModEximBucket() {
	url := "http://localhost:8000/api/admin/log-bucket/MOD_EXIM"
	// Create a new request using http.
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("[err][admin] creating request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Set custom admin auth header.
	req.Header.Set("Admin-Authorization", adminAuthToken)

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())
}
