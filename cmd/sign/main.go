package main

import (
	"crypto/rand"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func main() {
	host, err := newHost()
	if err != nil {
		panic(err)
	}

	fmt.Println("host ID:", host.ID())

	// --- crypto.PrivKey signing demo ---

	// Retrieve the private key from the peerstore
	privKey := host.Peerstore().PrivKey(host.ID())
	if privKey == nil {
		panic("no private key found in peerstore")
	}

	fmt.Printf("\n--- private key info ---\n")
	fmt.Printf("key type: %s\n", privKey.Type())
	raw, _ := privKey.Raw()
	fmt.Printf("raw bytes: %d bytes\n", len(raw))

	// Get the public key
	pubKey := privKey.GetPublic()
	fmt.Printf("public key type: %s\n", pubKey.Type())
	pubRaw, _ := pubKey.Raw()
	fmt.Printf("public raw bytes: %d bytes\n", len(pubRaw))

	// Sign a message
	message := []byte("hello libp2p — this is a signed message")
	signature, err := privKey.Sign(message)
	if err != nil {
		panic(fmt.Sprintf("sign failed: %v", err))
	}
	fmt.Printf("\nmessage: %q\n", message)
	fmt.Printf("signature: %d bytes\n", len(signature))
	fmt.Printf("signature hex: %x\n", signature)

	// Verify the signature using the public key
	valid, err := pubKey.Verify(message, signature)
	if err != nil {
		panic(fmt.Sprintf("verify failed: %v", err))
	}
	fmt.Printf("\nverification result: %v\n", valid)

	// Demonstrate that a tampered message fails verification
	tampered := []byte("hello libp2p — this is a TAMPERED message")
	valid, err = pubKey.Verify(tampered, signature)
	if err != nil {
		fmt.Printf("tampered verify error (expected): %v\n", err)
	} else {
		fmt.Printf("tampered verification (should be false): %v\n", valid)
	}

	// Demonstrate serialization round-trip
	fmt.Printf("\n--- key serialization ---\n")
	privBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		panic(fmt.Sprintf("marshal private key failed: %v", err))
	}
	fmt.Printf("marshalled private key: %d bytes\n", len(privBytes))

	privKey2, err := crypto.UnmarshalPrivateKey(privBytes)
	if err != nil {
		panic(fmt.Sprintf("unmarshal private key failed: %v", err))
	}
	fmt.Printf("unmarshalled key type: %s\n", privKey2.Type())

	// Verify the deserialized key produces the same signature
	sig2, _ := privKey2.Sign(message)
	valid2, _ := pubKey.Verify(message, sig2)
	fmt.Printf("round-trip verify: %v\n", valid2)

	// Generate a fresh ephemeral key pair (not from the host)
	fmt.Printf("\n--- ephemeral key generation ---\n")
	_, pubKey2, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("generate key failed: %v", err))
	}

	// The ephemeral key's public key should NOT verify the host's signature
	valid3, _ := pubKey2.Verify(message, signature)
	fmt.Printf("wrong key verify (should be false): %v\n", valid3)

	fmt.Println("\n--- signing demo complete ---")
}
