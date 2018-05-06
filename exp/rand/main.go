package main

import (
	"fmt"

	"GoBlog/rand"
)

func main() {
	fmt.Println(rand.String(10))
	fmt.Println(rand.RememberToken())
}
