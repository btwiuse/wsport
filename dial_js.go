//go:build js

package wsport

import (
	"net/http"

	"github.com/coder/websocket"
)

func dialOptions(http.Header) *websocket.DialOptions {
	return &websocket.DialOptions{}
}
