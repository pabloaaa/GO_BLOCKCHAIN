package main

import (
	"log"
	"net"
	"testing"
	"time"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

func startNode(node *Node, address string) {
	go node.Start([]byte(address))
	time.Sleep(1 * time.Second) // Czekamy, aby serwer się uruchomił
}

func sendWelcomeRequest(node *Node, address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	welcomeRequest := &pb.WelcomeRequest{
		Message: []byte(address),
	}

	err = encodeMessage(conn, "WelcomeRequest", welcomeRequest)
	if err != nil {
		log.Fatal(err)
	}
}

func requestLatestBlock(node *Node, address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	emptyMessage := &pb.Empty{}
	err = encodeMessage(conn, "GetLatestBlock", emptyMessage)
	if err != nil {
		log.Fatal(err)
	}
}

func generateBlocks(num int) []*Block {
	blocks := make([]*Block, num)
	for i := 0; i < num; i++ {
		block := &Block{
			Index:        uint64(i + 1), // Indeks zaczyna się od 1, ponieważ 0 to blok genesis
			Timestamp:    uint64(time.Now().Unix()),
			Transactions: make([]Transaction, 0),
			PreviousHash: []byte("0"),
			Data:         uint64(i),
		}
		if i > 0 {
			block.PreviousHash = blocks[i-1].calculateHash()
		}
		blocks[i] = block
	}
	return blocks
}

func TestNewNode(t *testing.T) {
	blockchain := NewBlockchain()

	node := NewNode(blockchain)

	if node.blockchain != blockchain {
		t.Errorf("Expected blockchain to be %v, but got %v", blockchain, node.blockchain)
	}

	if len(node.nodes) != 0 {
		t.Errorf("Expected nodes list to be empty, but got %d nodes", len(node.nodes))
	}
}

func TestNodeStart(t *testing.T) {
	// Tworzymy nowy blockchain
	blockchain := NewBlockchain()

	// Tworzymy nowy węzeł
	node := NewNode(blockchain)

	// Adres, na którym serwer będzie nasłuchiwał
	address := "127.0.0.1:8080"

	// Uruchamiamy serwer w osobnej gorutynie
	go node.Start([]byte(address))

	// Czekamy chwilę, aby serwer się uruchomił
	time.Sleep(1 * time.Second)

	// Próbujemy połączyć się z serwerem
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Sprawdzamy, czy połączenie zostało nawiązane
	if conn == nil {
		t.Fatalf("Expected to establish a connection, but got nil")
	}
}

func TestNodeSync(t *testing.T) {
	// Tworzymy dwa blockchainy
	blockchain1 := NewBlockchain()
	blockchain2 := NewBlockchain()

	// Dodajemy bloki do blockchainów
	for _, block := range generateBlocks(3) {
		blockchain1.AddBlock(blockchain1.root, block)
	}
	for _, block := range generateBlocks(7) {
		blockchain2.AddBlock(blockchain2.root, block)
	}

	// Tworzymy dwa nody
	node1 := NewNode(blockchain1)
	node2 := NewNode(blockchain2)

	// Adresy, na których serwery będą nasłuchiwać
	address1 := "127.0.0.1:8081"
	address2 := "127.0.0.1:8082"

	// Uruchamiamy serwery w osobnych gorutynach
	startNode(node1, address1)
	startNode(node2, address2)

	// Node1 wysyła zapytanie o najnowszy blok do Node2
	requestLatestBlock(node1, address2)

	// Czekamy chwilę, aby synchronizacja się zakończyła
	time.Sleep(2 * time.Second)

	// Sprawdzamy, czy oba nody mają teraz taką samą liczbę bloków
	if len(node1.blockchain.root.Childs) != len(node2.blockchain.root.Childs) {
		t.Errorf("Expected both nodes to have the same number of blocks, but got %d and %d", len(node1.blockchain.root.Childs), len(node2.blockchain.root.Childs))
	}
}
