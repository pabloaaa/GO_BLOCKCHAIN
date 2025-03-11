package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

var node *src.Node

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	// Pobierz flagi z argumentów
	port := flag.String("port", "50001", "port to listen on")
	httpPort := flag.String("httpPort", "60001", "HTTP port to listen on")
	bootstrapAddress := flag.String("bootstrapAddress", "50001", "port of the bootstrap node")
	logLevel := flag.String("logLevel", "info", "log level (info or debug)")
	flag.Parse()

	// Ustaw poziom logowania
	src.SetLogLevel(*logLevel)

	src.Info("server: Initializing blockchain")
	// Inicjalizacja blockchaina
	blockchain := src.NewBlockchain()

	portInt, _ := strconv.Atoi(*port)
	bootstrapPortInt, _ := strconv.Atoi(*bootstrapAddress)

	src.Info(fmt.Sprintf("server: Using port %s and HTTP port %s", *port, *httpPort))

	// Inicjalizacja tcpConnectionManager
	connectionManager := src.NewTcpConnectionManager(portInt)

	// Inicjalizacja tcpMessageSender
	tcpMessageSender := src.NewTCPSender(connectionManager)

	// Inicjalizacja noda
	node = src.NewNode(blockchain, portInt, tcpMessageSender, bootstrapPortInt)

	src.Info("server: Starting TCP server")
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
	router.POST("/digging", startDigging)

	// Uruchomienie serwera HTTP
	src.Info(fmt.Sprintf("server: HTTP server started on %s", *httpPort))
	if err := router.Run(":" + *httpPort); err != nil {
		src.Error(fmt.Sprintf("server: Failed to start HTTP server: %v", err))
	}
}

// syncNodes synchronizuje węzły blockchaina
func syncNodes(c *gin.Context) {
	otherNodeAddress := c.Query("address")
	if otherNodeAddress == "" {
		src.Error("server: Address query parameter is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "address query parameter is required"})
		return
	}

	// Extract the port from the address
	parts := strings.Split(otherNodeAddress, ":")
	if len(parts) != 2 {
		src.Error("server: Invalid address format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address format"})
		return
	}
	otherNodePort, err := strconv.Atoi(parts[1])
	if err != nil {
		src.Error("server: Invalid port")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid port"})
		return
	}

	src.Debug(fmt.Sprintf("server: Starting synchronization with node: %d from node: %d", otherNodePort, node.GetAddress()))

	err = node.SyncNodes(otherNodePort)
	if err != nil {
		src.Error(fmt.Sprintf("server: Failed to synchronize nodes: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to synchronize nodes"})
		return
	}

	src.Info(fmt.Sprintf("server: Synchronization with node %d complete", otherNodePort))

	c.JSON(http.StatusOK, gin.H{"message": "synchronization complete"})
}

// getStatus zwraca obecny stan blockchaina jako HTML
func getStatus(c *gin.Context) {
	src.Debug("server: Getting blockchain status")
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
	for port, conn := range node.GetPortMap() {
		status := "inactive"
		if conn != nil {
			status = "active"
		}
		html += fmt.Sprintf("<li>Port: %d, Status: %s</li>", port, status)
	}
	html += "</ul>"

	src.Info("server: Blockchain status retrieved successfully")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// startDigging uruchamia nieskończoną pętlę do ciągłego poszukiwania nowych bloków
func startDigging(c *gin.Context) {
	go func() {
		for {
			node.TryToFindNewBlock()
		}
	}()
	c.JSON(http.StatusOK, gin.H{"status": "Continuous block finding started"})
}
