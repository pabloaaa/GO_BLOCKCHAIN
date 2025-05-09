package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
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

func startNode(port, httpPort, bootstrapPort int, logLevel string) {
	go func() {
		mainWithArgs(port, httpPort, bootstrapPort, logLevel)
	}()
}

func mainWithArgs(port, httpPort, bootstrapPort int, logLevel string) {
	// Ustaw poziom logowania
	src.SetLogLevel(logLevel)

	// Initialize blockchain
	blockchain := src.NewBlockchain()

	// Initialize tcpConnectionManager
	connectionManager := src.NewTcpConnectionManager(port)

	// Initialize tcpMessageSender
	tcpMessageSender := src.NewTCPSender(connectionManager)

	// Initialize node
	node := src.NewNode(blockchain, port, tcpMessageSender, bootstrapPort)

	// Start TCP server
	go node.Start()

	// Initialize Gin router
	router := gin.Default()

	// Define endpoints
	router.POST("/sync", func(c *gin.Context) {
		otherNodeAddress := c.Query("address")
		if otherNodeAddress == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "address query parameter is required"})
			return
		}

		parts := strings.Split(otherNodeAddress, ":")
		if len(parts) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address format"})
			return
		}
		otherNodePort, err := strconv.Atoi(parts[1])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid port"})
			return
		}

		err = node.SyncNodes(otherNodePort)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to synchronize nodes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "synchronization complete"})
	})
	router.GET("/status", func(c *gin.Context) {
		var blocks []*types.Block
		node.GetBlockchain().TraverseTree(func(node *types.BlockNode) bool {
			blocks = append(blocks, node.Block)
			return false
		})

		html := "<html><head><title>Blockchain Status</title></head><body><h1>Blockchain Status</h1><ul>"
		for _, block := range blocks {
			hash := block.CalculateHash()
			html += fmt.Sprintf("<li>Index: %d, Timestamp: %d, Previous Hash: %x, Hash: %x, Transactions: %v, Data: %d, Checkpoint: %t</li>",
				block.Index, block.Timestamp, block.PreviousHash, hash, block.Transactions, block.Data, block.Checkpoint)
		}
		html += "</ul>"

		html += fmt.Sprintf("<h2>Node Address: %d</h2>", node.GetAddress())
		html += "<h2>Connected Nodes:</h2><ul>"
		for _, addr := range node.GetNodes() {
			html += fmt.Sprintf("<li>%d</li>", addr)
		}
		html += "</ul>"

		html += "<h2>Connection Manager Status:</h2>"
		html += "<h3>Port Map:</h3><ul>"
		for port, conn := range node.GetPortMap() {
			status := "inactive"
			if conn != nil {
				status = "active"
			}
			html += fmt.Sprintf("<li>Port: %d, Status: %s</li>", port, status)
		}
		html += "</ul>"

		html += "</body></html>"

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})
	router.POST("/find_new_block", func(c *gin.Context) {
		go node.TryToFindNewBlock()
		c.JSON(http.StatusOK, gin.H{"status": "Finding new block started"})
	})

	if err := router.Run(":" + strconv.Itoa(httpPort)); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func TestNodeWelcomeMessage(t *testing.T) {
	// Ensure ports are cleared before starting the test
	clearAllPorts()
	defer clearAllPorts() // Ensure ports are cleared after the test

	// Start three nodes with default log level (info)
	startNode(50001, 60001, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the first node to start

	startNode(50002, 60002, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the second node to start

	startNode(50003, 60003, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the third node to start

	// Wait for nodes to fully initialize
	time.Sleep(5 * time.Second)

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

	// Extract nodes from HTML
	nodes1, err := extractNodes(string(body1))
	assert.NoError(t, err)
	nodes2, err := extractNodes(string(body2))
	assert.NoError(t, err)
	nodes3, err := extractNodes(string(body3))
	assert.NoError(t, err)

	// Check if nodes are aware of each other
	assert.Contains(t, nodes1, "50002")
	assert.Contains(t, nodes1, "50003")
	assert.NotContains(t, nodes1, "50001")

	assert.Contains(t, nodes2, "50001")
	assert.NotContains(t, nodes2, "50002")

	assert.Contains(t, nodes3, "50001")
	assert.Contains(t, nodes3, "50002")
	assert.NotContains(t, nodes3, "50003")
}

func TestNodeSynchronization(t *testing.T) {
	// Ensure ports are cleared before starting the test
	clearAllPorts()
	defer clearAllPorts() // Ensure ports are cleared after the test

	// Start two nodes with default log level (info)
	startNode(50001, 60001, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the first node to start

	startNode(50002, 60002, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the second node to start

	// Create two new blocks on the first node
	for i := 0; i < 2; i++ {
		resp, err := http.Post("http://localhost:60001/find_new_block", "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(1 * time.Second) // Wait for the block to be created
	}

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

	// Check if both nodes have the same blockchain length
	assert.Equal(t, length1, length2)
}

func TestAddTenBlocksAndSync(t *testing.T) {
	// Ensure ports are cleared before starting the test
	clearAllPorts()
	defer clearAllPorts() // Ensure ports are cleared after the test

	// Start two nodes with default log level (info)
	startNode(50001, 60001, 50001, "debug")
	time.Sleep(1 * time.Second) // Wait for the first node to start

	startNode(50002, 60002, 50001, "debug")
	time.Sleep(1 * time.Second) // Wait for the second node to start

	// Create ten new blocks on the first node
	for i := 0; i < 10; i++ {
		resp, err := http.Post("http://localhost:60001/find_new_block", "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(1 * time.Second) // Wait for the block to be created
	}

	// Wait for a while to ensure blocks are created
	time.Sleep(10 * time.Second)

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

	// Check if both nodes have the same blockchain length
	assert.Equal(t, length1, length2)
}

func TestThreeNodesAddBlocksAndSync(t *testing.T) {
	// Ensure ports are cleared before starting the test
	clearAllPorts()
	defer clearAllPorts() // Ensure ports are cleared after the test

	// Start three nodes with default log level (info)
	startNode(50001, 60001, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the first node to start

	startNode(50002, 60002, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the second node to start

	startNode(50003, 60003, 50001, "info")
	time.Sleep(1 * time.Second) // Wait for the third node to start

	// Create ten new blocks on the first node
	for i := 0; i < 10; i++ {
		resp, err := http.Post("http://localhost:60001/find_new_block", "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(2 * time.Second) // Wait for the block to be created
	}

	// Wait for a while to ensure blocks are created
	time.Sleep(10 * time.Second)

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

	// Extract blockchain lengths from HTML
	length1, err := extractBlockchainLength(string(body1))
	assert.NoError(t, err)
	length2, err := extractBlockchainLength(string(body2))
	assert.NoError(t, err)
	length3, err := extractBlockchainLength(string(body3))
	assert.NoError(t, err)

	// Check if all nodes have the same blockchain length
	assert.Equal(t, length1, length2)
	assert.Equal(t, length1, length3)

	// Create three new blocks on the second node
	for i := 0; i < 3; i++ {
		resp, err := http.Post("http://localhost:60002/find_new_block", "application/json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(1 * time.Second) // Wait for the block to be created
	}

	// Synchronize the third node with the second node
	resp, err := http.Post("http://localhost:60003/sync?address=localhost:50002", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Wait for synchronization to complete
	time.Sleep(2 * time.Second)

	// Check the status of the second and third nodes
	status2, err = http.Get("http://localhost:60002/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status2.StatusCode)

	status3, err = http.Get("http://localhost:60003/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, status3.StatusCode)

	// Read the response bodies
	body2, err = ioutil.ReadAll(status2.Body)
	assert.NoError(t, err)
	body3, err = ioutil.ReadAll(status3.Body)
	assert.NoError(t, err)

	// Extract blockchain lengths from HTML
	length2, err = extractBlockchainLength(string(body2))
	assert.NoError(t, err)
	length3, err = extractBlockchainLength(string(body3))
	assert.NoError(t, err)

	// Check if the second and third nodes have the same blockchain length
	assert.Equal(t, length2, length3)
}
