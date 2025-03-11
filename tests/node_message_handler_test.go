package tests

import (
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/mocks"
	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	. "github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/stretchr/testify/mock"
)

func TestHandleNodeMessage_WelcomeRequest(t *testing.T) {
	mockSender := new(mocks.MockTcpMessageSender)
	nodes := []int{8081, 8082}
	handler := NewNodeMessageHandler(mockSender, &nodes, 8080)

	mockSender.On("SendMsgToAddress", 8081, mock.Anything).Return(nil)

	msg := &pb.NodeMessage{
		NodeMessageType: &pb.NodeMessage_WelcomeRequest{
			WelcomeRequest: &pb.WelcomeRequest{SenderAddress: []byte("8081")},
		},
	}

	handler.HandleNodeMessage(msg)

	mockSender.AssertExpectations(t)
}

func TestHandleNodeMessage_WelcomeResponse(t *testing.T) {
	mockSender := new(mocks.MockTcpMessageSender)
	nodes := []int{8081}
	handler := NewNodeMessageHandler(mockSender, &nodes, 8080)

	nodesAddresses := [][]byte{[]byte("8082"), []byte("8083")}
	msg := &pb.NodeMessage{
		NodeMessageType: &pb.NodeMessage_WelcomeResponse{
			WelcomeResponse: &pb.WelcomeResponse{NodeAdresses: nodesAddresses},
		},
	}

	handler.HandleNodeMessage(msg)

	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
}

func TestBroadcastAddress(t *testing.T) {
	mockSender := new(mocks.MockTcpMessageSender)
	nodes := []int{8081, 8082}
	handler := NewNodeMessageHandler(mockSender, &nodes, 8080)

	mockSender.On("SendMsgToAddress", 8081, mock.Anything).Return(nil)
	mockSender.On("SendMsgToAddress", 8082, mock.Anything).Return(nil)

	handler.BroadcastAddress(nodes, 8080)

	mockSender.AssertExpectations(t)
}

func TestSyncNodes(t *testing.T) {
	mockSender := new(mocks.MockTcpMessageSender)
	nodes := []int{8081}
	handler := NewNodeMessageHandler(mockSender, &nodes, 8080)

	mockSender.On("SendMsgToAddress", 8081, mock.Anything).Return(nil)

	err := handler.SyncNodes(8081)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	mockSender.AssertExpectations(t)
}
