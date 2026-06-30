package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "usage: box-label-pdf <input.xlsx> <invoice> <output.pdf>")
	os.Exit(2)
}
