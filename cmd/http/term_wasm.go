//go:build js && wasm

package main

import "net/http"

func termHandler() http.Handler {
	return http.NotFoundHandler()
}
