package main

import (
	"fmt"
	"os"
)

func main() {
	// Headless mode for testing/scripting: <input.xlsx> <invoice> <output.pdf>
	if args := os.Args[1:]; len(args) == 3 {
		res, err := generate(args[0], args[1], args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Printf("OK: %d branches, %d labels, %d pages -> %s\n",
			res.Branches, res.Labels, res.Pages, args[2])
		return
	}
	runGUI()
}
