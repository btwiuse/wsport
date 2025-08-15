//go:build !js

package wsport

import (
	"crypto/tls"
	"net/http"

	"github.com/coder/websocket"
)

func dialOptions(hdr http.Header) *websocket.DialOptions {
	return &websocket.DialOptions{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		HTTPHeader: hdr,
	}
}
