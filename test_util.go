package flyrpc

import "time"

type MockProtocol struct {
	*TcpProtocol
	packetChan chan *Packet
	delay      time.Duration
}

func NewMockProtocol() *MockProtocol {
	return &MockProtocol{nil, make(chan *Packet, 10), time.Millisecond}
	return NewMockDelayProtocol(time.Millisecond)
}

func NewMockDelayProtocol(delay time.Duration) *MockProtocol {
	return &MockProtocol{nil, make(chan *Packet, 10), delay}
}

func (mp *MockProtocol) SendPacket(pkt *Packet) error {
	go func() {
		<-time.After(mp.delay)
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
