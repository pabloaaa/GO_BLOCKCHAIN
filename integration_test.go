package main

import (
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

func clearAllPorts() {
	clearPort("50001")
	clearPort("50002")
	clearPort("50003")
	clearPort("60001")
	clearPort("60002")
	clearPort("60003")
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

func extractNodes(htmlBody string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	var nodes []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					nodes = append(nodes, c.Data)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return nodes, nil
}

func extractNodeAddress(htmlBody string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return "", err
	}

	var address string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			if strings.Contains(n.FirstChild.Data, "Node Address:") {
				address = strings.TrimSpace(strings.TrimPrefix(n.FirstChild.Data, "Node Address:"))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return address, nil
}

func TestNodeSynchronization(t *testing.T) {
	// Ensure ports are cleared
	clearAllPorts()

	// Defer killing processes on ports
	defer killProcessesOnPort("50001")
	defer killProcessesOnPort("50002")
	defer killProcessesOnPort("60001")
	defer killProcessesOnPort("60002")

	// Start the first node on port 50001
	cmd1 := exec.Command("go", "run", "server.go", "-port", "50001", "-httpPort", "60001")
	err := cmd1.Start()
	assert.NoError(t, err)
	defer cmd1.Process.Kill()

	time.Sleep(2 * time.Second) // Wait for the first node to start

	// Make 15 requests to find new blocks
	for i := 0; i < 15; i++ {
		resp, err := http.Post("http://localhost:60001/find_new_block", "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Start the second node on port 50002
	cmd2 := exec.Command("go", "run", "server.go", "-port", "50002", "-httpPort", "60002")
	err = cmd2.Start()
	assert.NoError(t, err)
	defer cmd2.Process.Kill()

	time.Sleep(2 * time.Second) // Wait for the second node to start

	// Synchronize the second node with the first node
	resp, err := http.Post("http://localhost:60002/sync?address=localhost:50001", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Wait for synchronization to complete
	time.Sleep(2 * time.Second)

	// Check the status of both nodes
	status1, err := http.Get("http://localhost:60001/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status1.StatusCode)

	status2, err := http.Get("http://localhost:60002/status")
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

func TestNodeWelcomeMessage(t *testing.T) {
	// Ensure ports are cleared
	clearAllPorts()

	// Defer killing processes on ports
	defer killProcessesOnPort("50001")
	defer killProcessesOnPort("50002")
	defer killProcessesOnPort("50003")
	defer killProcessesOnPort("60001")
	defer killProcessesOnPort("60002")
	defer killProcessesOnPort("60003")

	// Start the first node on port 50001
	cmd1 := exec.Command("go", "run", "server.go", "-port", "50001", "-httpPort", "60001")
	err := cmd1.Start()
	assert.NoError(t, err)
	defer cmd1.Process.Kill()

	time.Sleep(2 * time.Second) // Wait for the first node to start

	// Start the second node on port 50002
	cmd2 := exec.Command("go", "run", "server.go", "-port", "50002", "-httpPort", "60002")
	err = cmd2.Start()
	assert.NoError(t, err)
	defer cmd2.Process.Kill()

	time.Sleep(2 * time.Second) // Wait for the second node to start

	// Start the third node on port 50003
	cmd3 := exec.Command("go", "run", "server.go", "-port", "50003", "-httpPort", "60003")
	err = cmd3.Start()
	assert.NoError(t, err)
	defer cmd3.Process.Kill()

	time.Sleep(2 * time.Second) // Wait for the third node to start

	// Check the status of all nodes
	status1, err := http.Get("http://localhost:60001/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status1.StatusCode)

	status2, err := http.Get("http://localhost:60002/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status2.StatusCode)

	status3, err := http.Get("http://localhost:60003/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status3.StatusCode)

	// Read the response bodies
	body1, err := ioutil.ReadAll(status1.Body)
	assert.NoError(t, err)
	body2, err := ioutil.ReadAll(status2.Body)
	assert.NoError(t, err)
	body3, err := ioutil.ReadAll(status3.Body)
	assert.NoError(t, err)

	// Extract nodes and node addresses from HTML
	nodes1, err := extractNodes(string(body1))
	assert.NoError(t, err)
	nodes2, err := extractNodes(string(body2))
	assert.NoError(t, err)
	nodes3, err := extractNodes(string(body3))
	assert.NoError(t, err)

	address1, err := extractNodeAddress(string(body1))
	assert.NoError(t, err)
	address2, err := extractNodeAddress(string(body2))
	assert.NoError(t, err)
	address3, err := extractNodeAddress(string(body3))
	assert.NoError(t, err)

	// Check if nodes have the correct addresses and connections
	assert.Equal(t, "localhost:50001", address1)
	assert.Contains(t, nodes1, "localhost:50002")
	assert.Contains(t, nodes1, "localhost:50003")
	assert.NotContains(t, nodes1, "localhost:50001")

	assert.Equal(t, "localhost:50002", address2)
	assert.Contains(t, nodes2, "localhost:50001")
	assert.NotContains(t, nodes2, "localhost:50002")

	assert.Equal(t, "localhost:50003", address3)
	assert.Contains(t, nodes3, "localhost:50001")
	assert.Contains(t, nodes3, "localhost:50002")
	assert.NotContains(t, nodes3, "localhost:50003")
}
