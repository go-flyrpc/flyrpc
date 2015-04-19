package fly

type MockProtocol struct {
	packetChan chan *Packet
}

func NewMockProtocol() *MockProtocol {
	return &MockProtocol{
		packetChan: make(chan *Packet, 10),
	}
}

func (mp *MockProtocol) SendPacket(pkt *Packet) error {
	mp.packetChan <- pkt
	return nil
}

func (mp *MockProtocol) ReadPacket() (*Packet, error) {
	pkt := <-mp.packetChan
	return pkt, nil
}

func (mp *MockProtocol) Close() error {
	return nil
}
