package main

import (
	"fmt"
	"log"

	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/webpush"
)

func main() {
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Add these to your .env file:")
	fmt.Println()
	fmt.Printf("VAPID_PRIVATE_KEY=%s\n", priv)
	fmt.Printf("VAPID_PUBLIC_KEY=%s\n", pub)
}
