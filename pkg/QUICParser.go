package pkg

import (
	"encoding/binary"
	"errors"
)


type ClientHello struct {
	SNI  string
	ALPN []string
}


// ParseQUICClientHello extracts TLS ClientHello from QUIC Initial packet
func ParseQUICClientHello(pkt []byte) (*ClientHelloInfo, error) {

	if len(pkt) < 6 {
		return nil, errors.New("short quic packet")
	}


	// Long header check
	if pkt[0]&0x80 == 0 {
		return nil, errors.New("not quic long header")
	}


	// Initial packet only
	packetType := (pkt[0] >> 4) & 0x03

	if packetType != 0 {
		return nil, errors.New("not initial packet")
	}


	offset := 1


	// Version
	if len(pkt) < offset+4 {
		return nil, errors.New("missing version")
	}

	offset += 4


	// Destination Connection ID
	dcidLen := int(pkt[offset])
	offset++

	offset += dcidLen


	// Source Connection ID
	scidLen := int(pkt[offset])
	offset++

	offset += scidLen


	if offset >= len(pkt) {
		return nil, errors.New("bad cid")
	}


	// Token length
	tokenLen, n := quicVarInt(pkt[offset:])
	offset += n


	offset += int(tokenLen)


	if offset >= len(pkt) {
		return nil, errors.New("missing length")
	}


	// Packet length
	_, n = quicVarInt(pkt[offset:])
	offset += n


	// Packet number length
	pnLen := int((pkt[0] & 0x03) + 1)

	offset += pnLen


	if offset >= len(pkt) {
		return nil, errors.New("missing payload")
	}


	crypto, err := extractCryptoFrames(pkt[offset:])

	if err != nil {
		return nil, err
	}


	// crypto contains TLS handshake bytes
	return ParseClientHello(crypto)
}



func extractCryptoFrames(payload []byte) ([]byte,error) {

	offset := 0


	for offset < len(payload) {


		frameType, n := quicVarInt(payload[offset:])

		offset += n


		// CRYPTO frame type = 0x06
		if frameType == 0x06 {


			cryptoOffset, n :=
				quicVarInt(payload[offset:])

			offset += n


			length, n :=
				quicVarInt(payload[offset:])

			offset += n


			if int(length)+offset > len(payload) {
				return nil, errors.New("crypto overflow")
			}


			// Normally offset should be 0 for ClientHello
			if cryptoOffset != 0 {
				continue
			}


			return payload[offset:offset+int(length)], nil
		}


		// Skip unknown frame
		break
	}


	return nil, errors.New("crypto frame not found")
}



func quicVarInt(b []byte) (uint64,int) {


	if len(b) == 0 {
		return 0,0
	}


	prefix := b[0] >> 6


	switch prefix {


	case 0:
		return uint64(b[0]&0x3f),1


	case 1:
		if len(b)<2 {
			return 0,0
		}

		return uint64(binary.BigEndian.Uint16(b[:2])&0x3fff),2


	case 2:
		if len(b)<4 {
			return 0,0
		}

		return uint64(binary.BigEndian.Uint32(b[:4])&0x3fffffff),4


	case 3:
		if len(b)<8 {
			return 0,0
		}

		return binary.BigEndian.Uint64(b[:8])&0x3fffffffffffffff,8
	}


	return 0,0
}