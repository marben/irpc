package irpc

import (
	"fmt"
	"io"
	"sync"

	"github.com/marben/irpc/irpcgen"
)

type serializer struct {
	m   sync.Mutex
	enc *irpcgen.Encoder
}

func newSerializer(conn io.Writer) *serializer {
	return &serializer{enc: irpcgen.NewEncoder(conn)}
}

// returns ErrEndpointClosed if endpoint is in closing state
func (s *serializer) serializePacket(data ...irpcgen.Serializable) error {
	s.m.Lock()
	defer s.m.Unlock()

	return s.serializePacketLocked(data...)
}

func (s *serializer) serializePacketLocked(data ...irpcgen.Serializable) error {
	for _, d := range data {
		if err := d.Serialize(s.enc); err != nil {
			return fmt.Errorf("data.Serialize(): %w", err)
		}
	}

	if err := s.enc.Flush(); err != nil {
		return fmt.Errorf("encoder.Flush(): %w", err)
	}

	return nil
}

// respData is the actual serialized return data from the function
func (s *serializer) sendResponse(reqNum ReqNumT, respData irpcgen.Serializable) error {
	resp := responsePacket{
		ReqNum: reqNum,
	}

	header := packetHeader{
		typ: rpcResponsePacketType,
	}

	if err := s.serializePacket(header, resp, respData); err != nil {
		return fmt.Errorf("failed to serialize response to connection: %w", err)
	}

	return nil
}
