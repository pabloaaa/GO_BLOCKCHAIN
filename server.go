package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

var node *src.Node

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	log.Println("server: Initializing blockchain")
	// Inicjalizacja blockchaina
	blockchain := src.NewBlockchain()

	// Pobierz port z argumentów
	port := flag.String("port", "50001", "port to listen on")
	httpPort := flag.String("httpPort", "60001", "HTTP port to listen on")
	bootstrapAddress := flag.String("bootstrapAddress", "50001", "port of the bootstrap node")
	flag.Parse()

	portInt, _ := strconv.Atoi(*port)
	bootstrapPortInt, _ := strconv.Atoi(*bootstrapAddress)

	log.Printf("server: Using port %s and HTTP port %s", *port, *httpPort)

	// Inicjalizacja tcpConnectionManager
	connectionManager := src.NewTcpConnectionManager(portInt)

	// Inicjalizacja tcpMessageSender
	tcpMessageSender := src.NewTCPSender(connectionManager)

	// Inicjalizacja noda
	node = src.NewNode(blockchain, portInt, tcpMessageSender, bootstrapPortInt)

	log.Println("server: Starting TCP server")
	// Start TCP server
	go node.Start()

	// Inicjalizacja routera Gin
	router := gin.Default()

	// Definiowanie endpointów

	router.POST("/sync", syncNodes)
	router.GET("/status", getStatus)
	router.POST("/find_new_block", func(c *gin.Context) {
		go node.TryToFindNewBlock()
		c.JSON(http.StatusOK, gin.H{"status": "Finding new block started"})
	})

	// Uruchomienie serwera HTTP
	log.Printf("server: HTTP server started on %s", *httpPort)
	if err := router.Run(":" + *httpPort); err != nil {
		log.Fatalf("server: Failed to start HTTP server: %v", err)
	}
}

// syncNodes synchronizuje węzły blockchaina
func syncNodes(c *gin.Context) {
	otherNodeAddress := c.Query("address")
	if otherNodeAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address query parameter is required"})
		return
	}

	otherNodePort, _ := strconv.Atoi(otherNodeAddress)

	log.Printf("server: Starting synchronization with node: %d from node: %d", otherNodePort, node.GetAddress())

	err := node.SyncNodes(otherNodePort)
	if err != nil {
		log.Printf("server: Failed to synchronize nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to synchronize nodes"})
		return
	}

	log.Printf("server: Synchronization with node %d complete", otherNodePort)

	c.JSON(http.StatusOK, gin.H{"message": "synchronization complete"})
}

// getStatus zwraca obecny stan blockchaina jako HTML
func getStatus(c *gin.Context) {
	log.Println("server: Getting blockchain status")
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

	// Dodaj adres noda i adresy innych połączonych nodów
	html += fmt.Sprintf("<h2>Node Address: %d</h2>", node.GetAddress())
	html += "<h2>Connected Nodes:</h2><ul>"
	for _, addr := range node.GetNodes() {
		html += fmt.Sprintf("<li>%d</li>", addr)
	}
	html += "</ul>"

	// Dodaj informacje z TcpConnectionManager
	html += "<h2>Connection Manager Status:</h2>"
	html += "<h3>Port Map:</h3><ul>"
	for listeningPort, sendingPorts := range node.GetPortMap() {
		html += fmt.Sprintf("<li>Listening Port: %d<ul>", listeningPort)
		for sendingPort, conn := range sendingPorts {
			status := "inactive"
			if conn != nil {
				status = "active"
			}
			html += fmt.Sprintf("<li>Sending Port: %d, Status: %s</li>", sendingPort, status)
		}
		html += "</ul></li>"
	}
	html += "</ul>"

	html += "</body></html>"

	log.Println("server: Blockchain status retrieved successfully")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
