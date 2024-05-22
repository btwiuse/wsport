package main

import (
	"log"

	// We need to import libp2p's libraries that we use in this project.
	"github.com/btwiuse/gost"
	"github.com/libp2p/go-libp2p/core/network"
)

var autoHandler = gost.AutoHandler()

func AutoHandler(stream network.Stream) {
	// Remember to close the stream when we are done.
	c := &conn{stream}
	defer stream.Close()

	log.Println("Got a new stream!")
	autoHandler.Handle(c)
	log.Println("Stream handled!")
}
