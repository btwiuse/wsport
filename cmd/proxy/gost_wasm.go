//go:build js && wasm

package main

import (
	"io"
	"net"
)

type wasmHandler struct{}

func (wasmHandler) Init(opts ...interface{}) {}
func (wasmHandler) Handle(conn net.Conn)     { io.Copy(io.Discard, conn) }

// autoHandler returns a no-op handler for WASM builds.
// The real gost library uses syscalls not available on js/wasm.
func autoHandler() interface {
	Init(opts ...interface{})
	Handle(net.Conn)
} {
	return wasmHandler{}
}
