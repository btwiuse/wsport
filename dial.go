package wsport

import (
	"context"
	"net"
	"net/http"

	"github.com/coder/websocket"
)

func DialConn(ctx context.Context, addr string, hdr http.Header) (conn net.Conn, err error) {
	wsconn, _, err := websocket.Dial(
		ctx,
		addr,
		dialOptions(hdr),
	)
	if err != nil {
		return nil, err
	}

	return websocket.NetConn(context.Background(), wsconn, websocket.MessageBinary), nil
}
