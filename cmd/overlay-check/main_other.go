//go:build !windows || !amd64

package main

import "fmt"

func main() { fmt.Println("Overlay validation requires Windows x64.") }
