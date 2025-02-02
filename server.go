package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

var node *src.Node

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	// Inicjalizacja blockchaina
	blockchain := src.NewBlockchain()

	// Pobierz port z argumentów
	port := flag.String("port", "50001", "port to listen on")
	httpPort := flag.String("httpPort", "60001", "HTTP port to listen on")
	bootstrapAddress := flag.String("bootstrapAddress", "localhost:50001", "address of the bootstrap node")
	flag.Parse()

	// Inicjalizacja tcpMessageSender
	tcpMessageSender := src.NewTCPSender()

	// Inicjalizacja noda
	node = src.NewNode(blockchain, []byte("localhost:"+*port), tcpMessageSender, []byte(*bootstrapAddress))

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
	log.Printf("HTTP server started on %s", *httpPort)
	if err := router.Run(":" + *httpPort); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// syncNodes synchronizuje węzły blockchaina
func syncNodes(c *gin.Context) {
	otherNodeAddress := c.Query("address")
	if otherNodeAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address query parameter is required"})
		return
	}

	log.Printf("Starting synchronization with node: %s from node: %s", otherNodeAddress, node.GetAddress())

	err := node.SyncNodes([]byte(otherNodeAddress))
	if err != nil {
		log.Printf("Failed to synchronize nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to synchronize nodes"})
		return
	}

	log.Printf("Synchronization with node %s complete", otherNodeAddress)

	c.JSON(http.StatusOK, gin.H{"message": "synchronization complete"})
}

// getStatus zwraca obecny stan blockchaina jako HTML
func getStatus(c *gin.Context) {
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
	html += fmt.Sprintf("<h2>Node Address: %s</h2>", string(node.GetAddress()))
	html += "<h2>Connected Nodes:</h2><ul>"
	for _, addr := range node.GetNodes() {
		html += fmt.Sprintf("<li>%s</li>", string(addr))
	}
	html += "</ul></body></html>"

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
