package fly

import "time"

type MockProtocol struct {
	*TcpProtocol
	packetChan chan *Packet
}

func NewMockProtocol() *MockProtocol {
	return &MockProtocol{nil, make(chan *Packet, 10)}
}

func (mp *MockProtocol) SendPacket(pkt *Packet) error {
	go func() {
		time.After(time.Millisecond)
		mp.packetChan <- pkt
	}()
	return nil
}

func (mp *MockProtocol) ReadPacket() (*Packet, error) {
	pkt := <-mp.packetChan
	return pkt, nil
}

func (mp *MockProtocol) Close() error {
	return nil
}

type MockDeadProtocol struct {
	*MockProtocol
}

func NewMockDeadProtocol() *MockDeadProtocol {
	return &MockDeadProtocol{NewMockProtocol()}
}

func (p *MockDeadProtocol) SendPacket(pkt *Packet) error {
	return nil
}
