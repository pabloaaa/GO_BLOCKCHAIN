package main

import (
	"flag"
	"fmt"
	"html/template"
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

	// Pobierz port z argumentów lub użyj domyślnego
	port := flag.String("port", "8082", "port to listen on")
	flag.Parse()

	// Sprawdź, czy port jest już zajęty
	ln, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		// Jeśli port jest zajęty, użyj losowego dostępnego portu
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			fmt.Printf("Failed to find an available port: %v\n", err)
			return
		}
		*port = fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
		ln.Close()
	}

	fmt.Printf("Starting server on port %s\n", *port) // Debugowanie

	// Inicjalizacja noda
	node = src.NewNode(blockchain, ":"+*port)

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
	go node.Start()

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

	// Wywołaj metodę requestLatestBlock, aby zsynchronizować bloki
	node.GetMessageSender().SendMsg([]byte(fmt.Sprintf("http://%s/latestblock", otherNodeAddress)))

	c.JSON(http.StatusOK, gin.H{"message": "synchronization complete"})
}

// getStatus zwraca obecny stan blockchaina jako raport HTML
func getStatus(c *gin.Context) {
	var blocks []*types.Block
	node.GetBlockchain().TraverseTree(func(node *types.BlockNode) bool {
		blocks = append(blocks, node.Block)
		return false
	})

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Blockchain Status</title>
	</head>
	<body>
		<h1>Blockchain Status</h1>
		<ul>
			{{range .}}
			<li>
				<p>Index: {{.Index}}</p>
				<p>Timestamp: {{.Timestamp}}</p>
				<p>Previous Hash: {{.PreviousHash}}</p>
				<p>Hash: {{.CalculateHash}}</p>
				<p>Transactions:</p>
				<ul>
					{{range .Transactions}}
					<li>Sender: {{.Sender}}, Receiver: {{.Receiver}}, Amount: {{.Amount}}</li>
					{{end}}
				</ul>
				<p>Data: {{.Data}}</p>
				<p>Checkpoint: {{.Checkpoint}}</p>
			</li>
			{{end}}
		</ul>
	</body>
	</html>
	`

	t, err := template.New("status").Parse(tmpl)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error parsing template: %v", err)
		return
	}

	c.Header("Content-Type", "text/html")
	err = t.Execute(c.Writer, blocks)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error executing template: %v", err)
	}
}
