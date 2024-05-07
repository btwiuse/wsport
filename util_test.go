package wsport

import (
	"testing"

	ma "github.com/multiformats/go-multiaddr"
)

// addr := fmt.Sprintf("/dns/example.com/tcp/443/tls/sni/example.com/x-parity-wss/%%2F%d%%2F", port)
// addr := fmt.Sprintf("/dns/example.com/tcp/80/x-parity-ws/%%2F%d%%2F", port)
// addr := "/dns/example.com/tcp/8080/x-parity-ws/%2Fws"
// addr := fmt.Sprintf("/dns/example.com/tcp/443/tls/sni/example.com/x-parity-wss/%%2Fws%d", port)
// addr := fmt.Sprintf("/dns/example.com/tcp/443/x-parity-wss/%%2Fws%d", port)

func TestFromString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			"wss://example.com",
			"/dns/example.com/tcp/443/wss",
		},
		{
			"ws://example.com",
			"/dns/example.com/tcp/80/ws",
		},
		{
			"https://example.com",
			"/dns/example.com/tcp/443/wss",
		},
		{
			"http://example.com",
			"/dns/example.com/tcp/80/ws",
		},
		{
			"wss://example.com/ws",
			"/dns/example.com/tcp/443/x-parity-wss/%2Fws",
		},
		{
			"ws://example.com/ws",
			"/dns/example.com/tcp/80/x-parity-ws/%2Fws",
		},
		{
			"https://example.com/ws",
			"/dns/example.com/tcp/443/x-parity-wss/%2Fws",
		},
		{
			"http://example.com/ws",
			"/dns/example.com/tcp/80/x-parity-ws/%2Fws",
		},
		{
			"wss://example.com/ws/8080",
			"/dns/example.com/tcp/443/x-parity-wss/%2Fws%2F8080",
		},
		{
			"ws://example.com/ws/8080",
			"/dns/example.com/tcp/80/x-parity-ws/%2Fws%2F8080",
		},
		{
			"https://example.com:8080/ws",
			"/dns/example.com/tcp/8080/x-parity-wss/%2Fws",
		},
		{
			"http://example.com:8080/ws",
			"/dns/example.com/tcp/8080/x-parity-ws/%2Fws",
		},
	}

	for _, c := range cases {
		got, err := FromString(c.in)
		if err != nil {
			t.Errorf("FromString(%q) error: %v", c.in, err)
		}
		if !got.Equal(ma.StringCast(c.want)) {
			t.Errorf("FromString(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
