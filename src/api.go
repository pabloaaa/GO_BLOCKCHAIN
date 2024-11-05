package src

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

type API struct {
	node *Node
}

func NewAPI(blockchain interfaces.BlockchainInterface, address string) *API {
	node := NewNode(blockchain, address)
	return &API{node: node}
}

func (api *API) StartNode() {
	go api.node.Start()
}

func (api *API) DisplayBlockchain(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("blockchain").Parse(`
		<html>
		<head>
			<title>Blockchain State</title>
		</head>
		<body>
			<h1>Blockchain State</h1>
			{{range .}}
				<p>Index: {{.Index}}, Timestamp: {{.Timestamp}}, Data: {{.Data}}, Checkpoint: {{.Checkpoint}}</p>
			{{end}}
		</body>
		</html>
	`))

	var blocks []*types.Block
	api.node.traverseTree(func(node *types.BlockNode) bool {
		blocks = append(blocks, node.Block)
		return false
	})

	tmpl.Execute(w, blocks)
}

func (api *API) RequestLongestChain(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	messageSender, err := NewTCPSender(address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to node: %v", err), http.StatusInternalServerError)
		return
	}

	blockHandler := NewBlockMessageHandler(api.node.blockchain, messageSender)
	blockHandler.GetLatestBlock()

	fmt.Fprintf(w, "Requested longest chain from node at %s", address)
}

func (api *API) Serve() {
	http.HandleFunc("/display", api.DisplayBlockchain)
	http.HandleFunc("/request-longest-chain", api.RequestLongestChain)
	http.ListenAndServe(api.node.GetAddress(), nil)
}
