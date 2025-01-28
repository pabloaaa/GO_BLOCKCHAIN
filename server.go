package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

var node *src.Node

func main() {
	// Inicjalizacja blockchaina
	blockchain := src.NewBlockchain()

	// Pobierz port z argumentów
	port := flag.String("port", "0", "port to listen on")
	flag.Parse()

	// Sprawdź, czy port jest już zajęty lub ustawiony na 0
	ln, err := net.Listen("tcp", ":"+*port)
	if err != nil || *port == "0" {
		// Jeśli port jest zajęty lub ustawiony na 0, użyj losowego dostępnego portu
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			fmt.Printf("Failed to find an available port: %v\n", err)
			return
		}
		*port = fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
	}
	ln.Close()

	fmt.Printf("Starting server on port %s\n", *port) // Debugowanie

	// Inicjalizacja routera Gin
	router := gin.Default()

	// Definiowanie endpointów
	router.POST("/sync", syncNodes)
	router.GET("/status", getStatus)
	router.POST("/find_new_block", func(c *gin.Context) {
		go node.TryToFindNewBlock()
		c.JSON(http.StatusOK, gin.H{"status": "Finding new block started"})
	})

	// Uruchomienie serwera HTTP i wypisanie rzeczywistego portu
	server := &http.Server{
		Addr:    ":" + *port,
		Handler: router,
	}

	ln, err = net.Listen("tcp", server.Addr)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	fmt.Printf("Server is listening on port %d\n", ln.Addr().(*net.TCPAddr).Port)

	// Inicjalizacja noda
	node = src.NewNode(blockchain, "localhost:"+*port)

	go node.Start()

	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Failed to serve: %v\n", err)
	}
}

// syncNodes synchronizuje węzły blockchaina
func syncNodes(c *gin.Context) {
	otherNodeAddress := c.Query("address")
	if otherNodeAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address query parameter is required"})
		return
	}

	log.Printf("Starting synchronization with node: %s", otherNodeAddress)

	err := node.SyncNodes(otherNodeAddress)
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
	html += "</ul></body></html>"

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
