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
	labels := BuildLabels(branches, "IV-0001")
	fmt.Printf("%d branches, %d labels, %d pages\n",
		len(branches), len(labels), PageCount(len(labels)))
}
