package main

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func clearPort(port string) {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err == nil {
		conn.Close()
	}
}

func killProcessesOnPort(port string) {
	cmd := exec.Command("sh", "-c", "lsof -i :"+port+" | grep LISTEN | awk '{print $2}' | xargs kill -9")
	cmd.Run()
}

func extractBlockchainLength(htmlBody string) (int, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return 0, err
	}

	var f func(*html.Node)
	var length int
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			length++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return length, nil
}

func TestNodeSynchronization(t *testing.T) {
	// Ensure ports are cleared
	clearPort("49990")
	clearPort("49991")

	// Defer killing processes on ports 49990 and 49991
	defer killProcessesOnPort("49990")
	defer killProcessesOnPort("49991")

	// Start the first node on port 49990
	cmd1 := exec.Command("go", "run", "server.go", "-port", "49990", "-messageSenderAddress", "localhost:50000")
	stdout1, _ := cmd1.StdoutPipe()
	stderr1, _ := cmd1.StderrPipe()
	r1, w1 := io.Pipe()
	go func() {
		io.Copy(w1, stdout1)
	}()
	go func() {
		io.Copy(w1, stderr1)
	}()
	go func() {
		io.Copy(t.LogWriter(), r1)
	}()
	err := cmd1.Start()
	assert.NoError(t, err)
	defer cmd1.Process.Kill()

	time.Sleep(2 * time.Second) // Wait for the first node to start

	// Make 15 requests to find new blocks
	for i := 0; i < 15; i++ {
		resp, err := http.Post("http://localhost:49990/find_new_block", "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Start the second node on port 49991
	cmd2 := exec.Command("go", "run", "server.go", "-port", "49991", "-messageSenderAddress", "localhost:50000")
	stdout2, _ := cmd2.StdoutPipe()
	stderr2, _ := cmd2.StderrPipe()
	r2, w2 := io.Pipe()
	go func() {
		io.Copy(w2, stdout2)
	}()
	go func() {
		io.Copy(w2, stderr2)
	}()
	go func() {
		io.Copy(t.LogWriter(), r2)
	}()
	err = cmd2.Start()
	assert.NoError(t, err)
	defer cmd2.Process.Kill()

	time.Sleep(2 * time.Second) // Wait for the second node to start

	// Synchronize the second node with the first node
	resp, err := http.Post("http://localhost:49991/sync?address=localhost:49990", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Wait for synchronization to complete
	time.Sleep(2 * time.Second)

	// Check the status of both nodes
	status1, err := http.Get("http://localhost:49990/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status1.StatusCode)

	status2, err := http.Get("http://localhost:49991/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status2.StatusCode)

	// Read the response bodies
	body1, err := ioutil.ReadAll(status1.Body)
	assert.NoError(t, err)
	body2, err := ioutil.ReadAll(status2.Body)
	assert.NoError(t, err)

	// Extract blockchain lengths from HTML
	length1, err := extractBlockchainLength(string(body1))
	assert.NoError(t, err)
	length2, err := extractBlockchainLength(string(body2))
	assert.NoError(t, err)

	// Log the lengths of the blockchains
	t.Logf("Length of blockchain1: %d", length1)
	t.Logf("Length of blockchain2: %d", length2)

	// Compare the lengths of the blockchains
	assert.Equal(t, length1, length2)
}
