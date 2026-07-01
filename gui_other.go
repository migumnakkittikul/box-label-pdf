//go:build !windows

package main

import (
	"fmt"
	"os"
)

// On non-Windows there is no GUI build; the headless CLI path is used instead.
func runGUI() {
	fmt.Fprintln(os.Stderr, "usage: box-label-pdf <input.xlsx> <เลขที่ใบกำกับ> <output.pdf>")
	os.Exit(2)
}
