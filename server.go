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
	"github.com/pabloaaa/GO_BLOCKCHAIN/frontend"
	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

var node *src.Node
var chatHistory []string // Globalna historia czatu (niezależna od komunikacji Protobuf)

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

	// Start digging automatically
	go func() {
		for {
			node.TryToFindNewBlock()
		}
	}()

	// Inicjalizacja routera Gin
	router := gin.Default()

	// Definiowanie endpointów
	router.POST("/sync", syncNodes)
	router.GET("/status", getStatus)
	router.GET("/blockchain-data", getBlockchainData) // Dodaj ten endpoint
	router.Static("/frontend", "./frontend")
	router.POST("/find_new_block", func(c *gin.Context) {
		go node.TryToFindNewBlock()
		c.JSON(http.StatusOK, gin.H{"status": "Finding new block started"})
	})
	router.POST("/digging", startDigging)
	router.GET("/chat-history", getChatHistory)

	// Uruchomienie serwera HTTP
	src.Info(fmt.Sprintf("server: HTTP server started on %s", *httpPort))
	go func() {
		if err := router.Run(":" + *httpPort); err != nil {
			src.Error(fmt.Sprintf("server: Failed to start HTTP server: %v", err))
		}
	}()

	// Start the frontend application in the main goroutine
	frontend.StartFrontend(node)
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
		html += fmt.Sprintf("<li>Index: %d, Timestamp: %d, Previous Hash: %x, Hash: %x, Data: %d, Checkpoint: %t",
			block.Index, block.Timestamp, block.PreviousHash, hash, block.Data, block.Checkpoint)
		
		// Add messages if they exist
		if len(block.Messages) > 0 {
			html += "<br>Messages: <ul>"
			for _, msg := range block.Messages {
				html += fmt.Sprintf("<li>%s</li>", msg)
			}
			html += "</ul>"
		}
		html += "</li>"
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

// getBlockchainData zwraca dane blockchaina w formacie DOT
func getBlockchainData(c *gin.Context) {
	var blocks []*types.Block
	node.GetBlockchain().TraverseTree(func(node *types.BlockNode) bool {
		blocks = append(blocks, node.Block)
		return false
	})

	dot := "digraph G {\n"
	for _, block := range blocks {
		dot += fmt.Sprintf("\"%x\" [label=\"Index: %d\\nTimestamp: %d\\nData: %d\"];\n", block.CalculateHash(), block.Index, block.Timestamp, block.Data)
		if block.PreviousHash != nil {
			dot += fmt.Sprintf("\"%x\" -> \"%x\";\n", block.PreviousHash, block.CalculateHash())
		}
	}
	dot += "}"

	// Log the generated DOT data
	src.Debug(fmt.Sprintf("Generated DOT data: %s", dot))

	c.String(http.StatusOK, dot)
}

// uruchamia nieskończoną pętlę do ciągłego poszukiwania nowych bloków
func startDigging(c *gin.Context) {
	go func() {
		for {
			node.TryToFindNewBlock()
		}
	}()
	c.JSON(http.StatusOK, gin.H{"status": "Continuous block finding started"})
}

// broadcastMessage dodaje wiadomość do historii czatu
func broadcastMessage(message string) {
	chatHistory = append(chatHistory, message)
	src.Info(fmt.Sprintf("server: Message added to chat history: %s", message))
}

// getChatHistory zwraca wiadomości z blockchainu
func getChatHistory(c *gin.Context) {
	// Pobierz wiadomości z blockchainu
	var blockchainMessages []string
	node.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
		for _, msg := range blockNode.Block.Messages {
			blockchainMessages = append(blockchainMessages, msg)
		}
		return false
	})
	
	c.JSON(http.StatusOK, gin.H{"chatHistory": blockchainMessages})
}

