package src

type testSender struct {
	queue [][]byte
}

func NewTestSender() *testSender {
	return &testSender{queue: make([][]byte, 0)}
}

func (s *testSender) SendMsg(data []byte) error {
	s.queue = append(s.queue, data)
	return nil
}

func (s *testSender) GetQueue() [][]byte {
	return s.queue
}
