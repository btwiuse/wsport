package wsport

import (
	"net/url"

	"github.com/libp2p/go-libp2p"
	ma "github.com/multiformats/go-multiaddr"
)

// ListenAddrStrings configures libp2p to listen on the given (unparsed)
// ws:// or wss:// addresses.
func ListenAddrStrings(s ...string) libp2p.Option {
	return func(cfg *libp2p.Config) error {
		for _, addrstr := range s {
			a, err := FromString(addrstr)
			if err != nil {
				return err
			}
			cfg.ListenAddrs = append(cfg.ListenAddrs, a)
		}
		return nil
	}
}

func FromString(s string) (ma.Multiaddr, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return FromURL(u)
}
