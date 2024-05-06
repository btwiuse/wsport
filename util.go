package wsport

import (
	"net/url"

	"github.com/libp2p/go-libp2p"
)

// ListenAddrStrings configures libp2p to listen on the given (unparsed)
// ws:// or wss:// addresses.
func ListenAddrStrings(s ...string) libp2p.Option {
	return func(cfg *libp2p.Config) error {
		for _, addrstr := range s {
			u, err := url.Parse(addrstr)
			if err != nil {
				return err
			}
			a, err := FromURL(u)
			if err != nil {
				return err
			}
			cfg.ListenAddrs = append(cfg.ListenAddrs, a)
		}
		return nil
	}
}
