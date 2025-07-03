package frontend

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

func StartFrontend(node *src.Node) {
	a := app.New()
	w := a.NewWindow(fmt.Sprintf("Blockchain Node %d", node.GetAddress()))

	// Create a graffiti-style label
	graffitiLabel := canvas.NewText(fmt.Sprintf("Blockchain Node %d", node.GetAddress()), theme.PrimaryColor())
	graffitiLabel.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
	graffitiLabel.TextSize = 20

	// Create a container to display chat history
	chatHistory := container.NewVBox()
	chatHistoryScroll := container.NewVScroll(chatHistory)
	chatHistoryScroll.SetMinSize(fyne.NewSize(400, 400))

	// Create labels for reward and pending messages
	rewardLabel := widget.NewLabel(fmt.Sprintf("Reward: %d", node.GetBlockchain().GetReward()))
	rewardLabel.TextStyle = fyne.TextStyle{Bold: true}
	rewardLabel.Alignment = fyne.TextAlignCenter

	pendingLabel := widget.NewLabel("Pending messages: 0")
	pendingLabel.TextStyle = fyne.TextStyle{Italic: true}
	pendingLabel.Alignment = fyne.TextAlignCenter

	// Function to update the status labels periodically
	go func() {
		for range time.Tick(1 * time.Second) {
			reward := node.GetBlockchain().GetReward()
			rewardLabel.SetText(fmt.Sprintf("Reward: %d", reward))
			rewardLabel.Refresh()
			
			// Get current pending messages count
			pendingCount := node.GetPendingMessageCount()
			pendingLabel.SetText(fmt.Sprintf("Pending messages: %d", pendingCount))
			pendingLabel.Refresh()
		}
	}()

	// Add initial system message to chat history
	chatHistory.Add(widget.NewLabel(fmt.Sprintf("=== Node %d Chat Started ===", node.GetAddress())))
	chatHistory.Refresh()

	// Track blockchain messages count to detect new messages
	var lastMessageCount int

	// Function to update chat history from blockchain
	updateChatFromBlockchain := func() {
		var blockchainMessages []string
		node.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
			for _, msg := range blockNode.Block.Messages {
				blockchainMessages = append(blockchainMessages, msg)
			}
			return false
		})

		// Only update if there are new messages
		if len(blockchainMessages) != lastMessageCount {
			lastMessageCount = len(blockchainMessages)
			
			// Clear and rebuild chat history
			chatHistory.Objects = nil
			chatHistory.Add(widget.NewLabel(fmt.Sprintf("=== Node %d Chat Started ===", node.GetAddress())))
			
			if len(blockchainMessages) == 0 {
				chatHistory.Add(widget.NewLabel("No messages in blockchain yet..."))
			} else {
				for i, msg := range blockchainMessages {
					chatHistory.Add(widget.NewLabel(fmt.Sprintf("[Block %d] %s", i+1, msg)))
				}
			}
			chatHistory.Refresh()
			
			// Auto-scroll to bottom
			if len(chatHistory.Objects) > 0 {
				chatHistoryScroll.ScrollToBottom()
			}
		}
	}

	// Real-time blockchain monitoring for new messages
	go func() {
		for range time.Tick(500 * time.Millisecond) { // Check every 500ms for new blocks
			updateChatFromBlockchain()
		}
	}()

	// Initial population of chat history
	updateChatFromBlockchain()

	// Create a message entry
	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("Type your message and press Enter or click Send...")
	
	// Function to send message (reusable)
	sendMessage := func() {
		message := messageEntry.Text
		if message != "" {
			// Check if node has enough reward
			if node.GetBlockchain().GetReward() < 10 {
				dialog.ShowInformation("Insufficient Reward", "You need at least 10 reward to send messages", w)
				return
			}

			// Deduct reward for sending message
			err := node.GetBlockchain().RewardNode(node.GetAddress(), -10)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to deduct reward: %v", err), w)
				return
			}

			// Add message to node's pending queue
			formattedMessage := fmt.Sprintf("NODE %d said: %s", node.GetAddress(), message)
			node.AddPendingMessage(formattedMessage)
			
			// Clear the input
			messageEntry.SetText("")
			
			// Show confirmation
			dialog.ShowInformation("Message Queued", 
				fmt.Sprintf("Message queued for next block. Reward: %d (-10)", node.GetBlockchain().GetReward()), w)
		}
	}
	
	// Add Enter key support
	messageEntry.OnSubmitted = func(text string) {
		sendMessage()
	}

	// Send button uses the same function
	sendButton := widget.NewButton("Send", sendMessage)

	// Create status container for reward and pending info
	statusContainer := container.NewHBox(rewardLabel, pendingLabel)

	// Create a main container with a vertical layout
	mainContainer := container.NewVBox(
		graffitiLabel,
		chatHistoryScroll,
		messageEntry,
		sendButton,
		statusContainer,
	)

	// Set the content of the window
	w.SetContent(mainContainer)

	// Resize the window
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
