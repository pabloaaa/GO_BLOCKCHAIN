package tests

import (
	"reflect"
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
)

func TestNewTestSender(t *testing.T) {
	sender := src.NewTestSender()
	
	if sender == nil {
		t.Error("NewTestSender returned nil")
	}
}

func TestTestSender_SendMsg(t *testing.T) {
	sender := src.NewTestSender()
	testData := []byte("test message")
	
	err := sender.SendMsg(testData)
	if err != nil {
		t.Errorf("SendMsg failed: %v", err)
	}
	
	queue := sender.GetQueue()
	if len(queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(queue))
	}
	
	if !reflect.DeepEqual(queue[0], testData) {
		t.Errorf("Expected data %v, got %v", testData, queue[0])
	}
}

func TestTestSender_SendMsg_Multiple(t *testing.T) {
	sender := src.NewTestSender()
	testData1 := []byte("first message")
	testData2 := []byte("second message")
	testData3 := []byte("third message")
	
	err := sender.SendMsg(testData1)
	if err != nil {
		t.Errorf("SendMsg failed: %v", err)
	}
	
	err = sender.SendMsg(testData2)
	if err != nil {
		t.Errorf("SendMsg failed: %v", err)
	}
	
	err = sender.SendMsg(testData3)
	if err != nil {
		t.Errorf("SendMsg failed: %v", err)
	}
	
	queue := sender.GetQueue()
	if len(queue) != 3 {
		t.Errorf("Expected queue length 3, got %d", len(queue))
	}
	
	if !reflect.DeepEqual(queue[0], testData1) {
		t.Errorf("Expected first message %v, got %v", testData1, queue[0])
	}
	
	if !reflect.DeepEqual(queue[1], testData2) {
		t.Errorf("Expected second message %v, got %v", testData2, queue[1])
	}
	
	if !reflect.DeepEqual(queue[2], testData3) {
		t.Errorf("Expected third message %v, got %v", testData3, queue[2])
	}
}

func TestTestSender_GetQueue_Empty(t *testing.T) {
	sender := src.NewTestSender()
	
	queue := sender.GetQueue()
	if queue == nil {
		t.Error("GetQueue should not return nil")
	}
	
	if len(queue) != 0 {
		t.Errorf("Expected empty queue, got length %d", len(queue))
	}
}

func TestTestSender_SendMsg_EmptyData(t *testing.T) {
	sender := src.NewTestSender()
	emptyData := []byte{}
	
	err := sender.SendMsg(emptyData)
	if err != nil {
		t.Errorf("SendMsg failed with empty data: %v", err)
	}
	
	queue := sender.GetQueue()
	if len(queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(queue))
	}
	
	if !reflect.DeepEqual(queue[0], emptyData) {
		t.Errorf("Expected empty data %v, got %v", emptyData, queue[0])
	}
}

func TestTestSender_SendMsg_NilData(t *testing.T) {
	sender := src.NewTestSender()
	var nilData []byte = nil
	
	err := sender.SendMsg(nilData)
	if err != nil {
		t.Errorf("SendMsg failed with nil data: %v", err)
	}
	
	queue := sender.GetQueue()
	if len(queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(queue))
	}
	
	if queue[0] != nil {
		t.Errorf("Expected nil data, got %v", queue[0])
	}
}