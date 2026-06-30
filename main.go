package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: box-label-pdf <input.xlsx> <invoice> <output.pdf>")
		os.Exit(2)
	}
	res, err := generate(args[0], args[1], args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("OK: %d branches, %d labels, %d pages -> %s\n",
		res.Branches, res.Labels, res.Pages, args[2])
}
