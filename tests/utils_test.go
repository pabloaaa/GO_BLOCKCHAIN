package tests

import (
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

func TestEncodeMessage(t *testing.T) {
	message := &block_chain.WelcomeRequest{
		SenderAddress: []byte("50001"),
	}

	data, err := src.EncodeMessage(message)
	if err != nil {
		t.Errorf("EncodeMessage failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("EncodeMessage returned empty data")
	}
}

func TestDecodeMessage(t *testing.T) {
	originalMessage := &block_chain.WelcomeRequest{
		SenderAddress: []byte("50001"),
	}

	data, err := proto.Marshal(originalMessage)
	if err != nil {
		t.Fatalf("Failed to marshal test message: %v", err)
	}

	decodedMessage := &block_chain.WelcomeRequest{}
	err = src.DecodeMessage(data, decodedMessage)
	if err != nil {
		t.Errorf("DecodeMessage failed: %v", err)
	}

	if string(decodedMessage.SenderAddress) != string(originalMessage.SenderAddress) {
		t.Errorf("Expected SenderAddress %s, got %s", string(originalMessage.SenderAddress), string(decodedMessage.SenderAddress))
	}
}

func TestPrepareProtoMessageToSend_WelcomeRequest(t *testing.T) {
	factory := src.NewMessageFactory()
	message := &block_chain.WelcomeRequest{
		SenderAddress: []byte("50001"),
	}

	data, err := src.PrepareProtoMessageToSend(factory, message)
	if err != nil {
		t.Errorf("PrepareProtoMessageToSend failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("PrepareProtoMessageToSend returned empty data")
	}
}

func TestPrepareProtoMessageToSend_WelcomeResponse(t *testing.T) {
	factory := src.NewMessageFactory()
	message := &block_chain.WelcomeResponse{
		NodeAdresses: [][]byte{[]byte("50002")},
	}

	data, err := src.PrepareProtoMessageToSend(factory, message)
	if err != nil {
		t.Errorf("PrepareProtoMessageToSend failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("PrepareProtoMessageToSend returned empty data")
	}
}

func TestPrepareProtoMessageToSend_BlockResponse(t *testing.T) {
	factory := src.NewMessageFactory()
	message := &block_chain.BlockResponse{
		Block: &block_chain.Block{Index: 1, Data: 100},
		SenderAddress: []byte("50001"),
	}

	data, err := src.PrepareProtoMessageToSend(factory, message)
	if err != nil {
		t.Errorf("PrepareProtoMessageToSend failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("PrepareProtoMessageToSend returned empty data")
	}
}

func TestPrepareProtoMessageToSend_BlockchainSyncRequest(t *testing.T) {
	factory := src.NewMessageFactory()
	message := &block_chain.BlockchainSyncRequest{
		SenderAddress: []byte("50001"),
	}

	data, err := src.PrepareProtoMessageToSend(factory, message)
	if err != nil {
		t.Errorf("PrepareProtoMessageToSend failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("PrepareProtoMessageToSend returned empty data")
	}
}

func TestPrepareProtoMessageToSend_ApprovedBlock(t *testing.T) {
	factory := src.NewMessageFactory()
	message := &block_chain.ApprovedBlock{
		Block: &block_chain.Block{
			Index: 1,
			Data:  100,
		},
	}

	data, err := src.PrepareProtoMessageToSend(factory, message)
	if err != nil {
		t.Errorf("PrepareProtoMessageToSend failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("PrepareProtoMessageToSend returned empty data")
	}
}

func TestPrepareProtoMessageToSend_BlockRequest(t *testing.T) {
	factory := src.NewMessageFactory()
	message := &block_chain.BlockRequest{
		SenderAddress: []byte("50001"),
	}

	data, err := src.PrepareProtoMessageToSend(factory, message)
	if err != nil {
		t.Errorf("PrepareProtoMessageToSend failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("PrepareProtoMessageToSend returned empty data")
	}
}


func TestEncodeDecodeMessage_RoundTrip(t *testing.T) {
	originalMessage := &block_chain.WelcomeRequest{
		SenderAddress: []byte("50001"),
	}

	encodedData, err := src.EncodeMessage(originalMessage)
	if err != nil {
		t.Fatalf("EncodeMessage failed: %v", err)
	}

	decodedMessage := &block_chain.WelcomeRequest{}
	err = src.DecodeMessage(encodedData, decodedMessage)
	if err != nil {
		t.Fatalf("DecodeMessage failed: %v", err)
	}

	if string(originalMessage.SenderAddress) != string(decodedMessage.SenderAddress) {
		t.Errorf("Round-trip encoding/decoding failed. Original: %s, Decoded: %s", string(originalMessage.SenderAddress), string(decodedMessage.SenderAddress))
	}
}