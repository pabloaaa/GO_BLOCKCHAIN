package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
)

func main() {
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	address := serverCmd.String("address", ":8080", "Node address")
	clientAddress := clientCmd.String("address", "", "Node address to query")

	if len(os.Args) < 2 {
		fmt.Println("expected 'server' or 'client' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverCmd.Parse(os.Args[2:])
		blockchain := src.NewBlockchain()
		api := src.NewAPI(blockchain, *address)
		api.StartNode()
		api.Serve()
	case "client":
		clientCmd.Parse(os.Args[2:])
		if *clientAddress == "" {
			fmt.Println("expected node address to query")
			os.Exit(1)
		}
		resp, err := http.Get(fmt.Sprintf("http://%s/display", *clientAddress))
		if err != nil {
			fmt.Printf("Failed to get blockchain from node at %s: %v\n", *clientAddress, err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %v\n", err)
			return
		}

		file, err := os.Create("blockchain.html")
		if err != nil {
			fmt.Printf("Failed to create HTML file: %v\n", err)
			return
		}
		defer file.Close()

		_, err = file.Write(body)
		if err != nil {
			fmt.Printf("Failed to write to HTML file: %v\n", err)
			return
		}

		fmt.Println("Blockchain HTML report generated: blockchain.html")

		// Open the HTML file in the default browser
		var cmd *exec.Cmd
		switch os := runtime.GOOS; os {
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", "blockchain.html")
		case "linux":
			cmd = exec.Command("xdg-open", "blockchain.html")
		case "darwin": // macOS
			cmd = exec.Command("open", "blockchain.html")
		default:
			fmt.Println("unsupported platform")
			return
		}
		cmd.Start()
	default:
		fmt.Println("expected 'server' or 'client' subcommands")
		os.Exit(1)
	}
}
