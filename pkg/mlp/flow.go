package mlp

type StreamFlowControl struct {
	MaxStreamData uint64
	Sent          uint64
	Received      uint64
}

func NewStreamFlowControl(credit uint64) *StreamFlowControl {
	return &StreamFlowControl{MaxStreamData: credit}
}

func (s *StreamFlowControl) Available() uint64     { return s.MaxStreamData - s.Sent }
func (s *StreamFlowControl) CanSend(n uint64) bool { return n <= s.Available() }
func (s *StreamFlowControl) OnSend(n uint64)       { s.Sent += n }
func (s *StreamFlowControl) UpdateCredit(n uint64) {
	if n > s.MaxStreamData {
		s.MaxStreamData = n
	}
}

type ConnectionFlowControl struct {
	MaxData  uint64
	Sent     uint64
	Received uint64
}

func NewConnectionFlowControl(credit uint64) *ConnectionFlowControl {
	return &ConnectionFlowControl{MaxData: credit}
}

func (c *ConnectionFlowControl) Available() uint64     { return c.MaxData - c.Sent }
func (c *ConnectionFlowControl) CanSend(n uint64) bool { return n <= c.Available() }
func (c *ConnectionFlowControl) OnSend(n uint64)       { c.Sent += n }
func (c *ConnectionFlowControl) UpdateCredit(n uint64) {
	if n > c.MaxData {
		c.MaxData = n
	}
}

type FlowControl struct {
	Connection    *ConnectionFlowControl
	Streams       map[uint32]*StreamFlowControl
	DefaultCredit uint64
}

func NewFlowControl(connCredit, streamCredit uint64) *FlowControl {
	return &FlowControl{Connection: NewConnectionFlowControl(connCredit), Streams: map[uint32]*StreamFlowControl{}, DefaultCredit: streamCredit}
}

func (fc *FlowControl) Stream(id uint32) *StreamFlowControl {
	if s, ok := fc.Streams[id]; ok {
		return s
	}
	s := NewStreamFlowControl(fc.DefaultCredit)
	fc.Streams[id] = s
	return s
}

func (fc *FlowControl) CanSend(id uint32, n uint64) bool {
	if !fc.Connection.CanSend(n) {
		return false
	}
	return fc.Stream(id).CanSend(n)
}

func (fc *FlowControl) OnSend(id uint32, n uint64) {
	fc.Connection.OnSend(n)
	fc.Stream(id).OnSend(n)
}
