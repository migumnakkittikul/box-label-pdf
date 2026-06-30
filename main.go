package main

import (
	"fmt"
	"os"
)

func main() {
	_, branches, err := ReadBranches("sample.xlsx")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("%d branches\n", len(branches))
}
