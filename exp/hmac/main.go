package main

import (
	"fmt"

	"coverd/hash"
)

func main() {
	hmac := hash.NewHMAC("my-secret-key")
	// This should print out:
	// 4waUFc1cnuxoM2oUOJfpGZLGP1asj35y7teuweSFgPY=
	fmt.Println(hmac.Hash("this is my string to hash"))
}
