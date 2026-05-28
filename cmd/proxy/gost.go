//go:build !js || !wasm

package main

import "github.com/btwiuse/gost"

func autoHandler() gost.Handler {
	return gost.AutoHandler()
}
