package pkg

import "errors"

type ClientHelloInfo struct {
	SNI  string
	ALPN []string
}

var (
	ErrNotTLS         = errors.New("not tls")
	ErrNotClientHello = errors.New("not client hello")
	ErrTruncated      = errors.New("truncated message")
)

func ParseClientHello(data []byte) (*ClientHelloInfo, error) {
	info := &ClientHelloInfo{}

	// TLS record header
	if len(data) < 5 {
		return nil, ErrTruncated
	}

	if data[0] != 0x16 {
		return nil, ErrNotTLS
	}

	recordLen := int(data[3])<<8 | int(data[4])

	if len(data) < 5+recordLen {
		return nil, ErrTruncated
	}

	data = data[5 : 5+recordLen]

	// Handshake header
	if len(data) < 4 {
		return nil, ErrTruncated
	}

	if data[0] != 0x01 {
		return nil, ErrNotClientHello
	}

	handshakeLen :=
		int(data[1])<<16 |
			int(data[2])<<8 |
			int(data[3])

	if len(data) < 4+handshakeLen {
		return nil, ErrTruncated
	}

	data = data[4 : 4+handshakeLen]

	// legacy_version + random
	if len(data) < 34 {
		return nil, ErrTruncated
	}

	data = data[34:]

	// session id
	var ok bool

	data, ok = skipVector8(data)
	if !ok {
		return nil, ErrTruncated
	}

	// cipher suites
	data, ok = skipVector16(data)
	if !ok {
		return nil, ErrTruncated
	}

	// compression methods
	data, ok = skipVector8(data)
	if !ok {
		return nil, ErrTruncated
	}

	// extensions
	extensions, rest, ok := readVector16(data)
	if !ok || len(rest) != 0 {
		return nil, ErrTruncated
	}

	for len(extensions) > 0 {
		if len(extensions) < 4 {
			return nil, ErrTruncated
		}

		typ := uint16(extensions[0])<<8 | uint16(extensions[1])
		l := int(extensions[2])<<8 | int(extensions[3])

		if len(extensions) < 4+l {
			return nil, ErrTruncated
		}

		ext := extensions[4 : 4+l]

		switch typ {

		// server_name
		case 0:
			if sni, ok := parseSNI(ext); ok {
				info.SNI = sni
			}

		// ALPN
		case 16:
			info.ALPN = parseALPN(ext)
		}

		extensions = extensions[4+l:]
	}

	return info, nil
}

func parseSNI(ext []byte) (string, bool) {
	names, _, ok := readVector16(ext)
	if !ok {
		return "", false
	}

	for len(names) > 0 {
		if len(names) < 3 {
			return "", false
		}

		nameType := names[0]

		nameLen := int(names[1])<<8 | int(names[2])

		if len(names) < 3+nameLen {
			return "", false
		}

		if nameType == 0 {
			return string(names[3 : 3+nameLen]), true
		}

		names = names[3+nameLen:]
	}

	return "", false
}

func parseALPN(ext []byte) []string {
	var result []string

	protos, _, ok := readVector16(ext)
	if !ok {
		return nil
	}

	for len(protos) > 0 {
		if len(protos) < 1 {
			return result
		}

		l := int(protos[0])

		if len(protos) < 1+l {
			return result
		}

		result = append(result, string(protos[1:1+l]))

		protos = protos[1+l:]
	}

	return result
}

func skipVector8(data []byte) ([]byte, bool) {
	if len(data) < 1 {
		return nil, false
	}

	l := int(data[0])

	if len(data) < 1+l {
		return nil, false
	}

	return data[1+l:], true
}

func skipVector16(data []byte) ([]byte, bool) {
	if len(data) < 2 {
		return nil, false
	}

	l := int(data[0])<<8 | int(data[1])

	if len(data) < 2+l {
		return nil, false
	}

	return data[2+l:], true
}

func readVector16(data []byte) (value []byte, rest []byte, ok bool) {
	if len(data) < 2 {
		return nil, nil, false
	}

	l := int(data[0])<<8 | int(data[1])

	if len(data) < 2+l {
		return nil, nil, false
	}

	return data[2 : 2+l], data[2+l:], true
}