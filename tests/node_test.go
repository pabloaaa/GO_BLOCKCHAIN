package tests

import (
	"log"
	"net"
	"testing"
	"time"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	. "github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

func startNode(node *Node, address string) {
	go node.Start()
	time.Sleep(1 * time.Second)
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

	data, err := EncodeMessage(welcomeRequest)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write(data)
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

	latestBlockRequest := &pb.GetLatestBlockRequest{}
	data, err := EncodeMessage(latestBlockRequest)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func generateBlocks(num int) []*types.Block {
	blocks := make([]*types.Block, num)
	for i := 0; i < num; i++ {
		block := &types.Block{
			Index:        uint64(i + 1), // Indeks zaczyna się od 1, ponieważ 0 to blok genesis
			Timestamp:    uint64(time.Now().Unix()),
			Transactions: make([]types.Transaction, 0),
			PreviousHash: []byte("0"),
			Data:         uint64(i),
		}
		if i > 0 {
			block.PreviousHash = blocks[i-1].CalculateHash()
		}
		blocks[i] = block
	}
	return blocks
}

func TestNewNode(t *testing.T) {
	blockchain := NewBlockchain()
	address := "127.0.0.1:8080"

	node := NewNode(blockchain, address)

	if node.GetBlockchain() != blockchain {
		t.Errorf("Expected blockchain to be %v, but got %v", blockchain, node.GetBlockchain())
	}

	if len(node.GetNodes()) != 0 {
		t.Errorf("Expected nodes list to be empty, but got %d nodes", len(node.GetNodes()))
	}
}

func TestNodeStart(t *testing.T) {
	// Tworzymy nowy blockchain
	blockchain := NewBlockchain()
	address := "127.0.0.1:8080"

	// Tworzymy nowy węzeł
	node := NewNode(blockchain, address)

	// Uruchamiamy serwer w osobnej gorutynie
	go node.Start()

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
		blockchain1.AddBlock(blockchain1.GetRoot(), block)
	}
	for _, block := range generateBlocks(7) {
		blockchain2.AddBlock(blockchain2.GetRoot(), block)
	}

	// Tworzymy dwa nody
	address1 := "127.0.0.1:8081"
	address2 := "127.0.0.1:8082"
	node1 := NewNode(blockchain1, address1)
	node2 := NewNode(blockchain2, address2)

	// Uruchamiamy serwery w osobnych gorutynach
	startNode(node1, address1)
	startNode(node2, address2)

	// Node1 wysyła zapytanie o najnowszy blok do Node2
	requestLatestBlock(node1, address2)

	// Czekamy chwilę, aby synchronizacja się zakończyła
	time.Sleep(2 * time.Second)

	// Sprawdzamy, czy oba nody mają teraz taką samą liczbę bloków
	log.Printf("Node1 blocks: %d, Node2 blocks: %d", len(node1.GetBlockchain().GetRoot().Childs), len(node2.GetBlockchain().GetRoot().Childs))
	if len(node1.GetBlockchain().GetRoot().Childs) != len(node2.GetBlockchain().GetRoot().Childs) {
		t.Errorf("Expected both nodes to have the same number of blocks, but got %d and %d", len(node1.GetBlockchain().GetRoot().Childs), len(node2.GetBlockchain().GetRoot().Childs))
	}
}
