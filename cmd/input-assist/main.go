// gemu-input is the independent, per-user desktop input companion.
package main

import (
	"kyberix.com/gemu/internal/assist"
	"log"
)

func main() {
	if err := assist.Run(); err != nil {
		log.Fatal(err)
	}
}
