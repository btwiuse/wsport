package main

import (
	"log"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/btwiuse/gost"
	"github.com/libp2p/go-libp2p/core/network"
)

var autoHandler = gost.AutoHandler()

// streamHandler is our function to handle any libp2p-net streams that belong
// to our protocol. The streams should contain an HTTP request which we need
// to parse, make on behalf of the original node, and then write the response
// on the stream, before closing it.
func connectHandler(stream network.Stream) {
	// Remember to close the stream when we are done.
	c := &conn{stream}
	defer stream.Close()

	log.Println("Got a new stream!")
	autoHandler.Handle(c)
	log.Println("Stream handled!")
}
