package config 

import "github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"


func buildMatcher(
	protocol inspector.Protocol,
	r matchInput,
) Matcher {


	switch protocol {


	case inspector.TLSProto:

		return &TLSMatcher{

			SNI: r.host,

			ALPN: r.alpn,
		}



	case inspector.PostgresProto:

		return &PostgresMatcher{

			Params: r.meta,
		}



	case inspector.HTTPProto:

		return &HTTPMatcher{

			Host: r.host,
		}



	default:

		return &TCPMatcher{}
	}
}


func parseProto (proto string) inspector.Protocol {
    var p inspector.Protocol
    switch proto {
    case "http":
        p = inspector.HTTPProto
    case "tls":
        p = inspector.TLSProto
    case "postgres":
        p = inspector.PostgresProto
    default: 
        p = inspector.RawTCPProto
    }

    return p
}