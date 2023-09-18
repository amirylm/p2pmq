package main

import (
	"fmt"

	"github.com/amirylm/p2pmq/commons"
)

func main() {
	_, skB64, err := commons.GetOrGeneratePrivateKey("")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", skB64)
}
