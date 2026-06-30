package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

// allocationSheet is the sheet name in sample.xlsx-style files that holds the
// per-branch breakdown of a purchase order.
const allocationSheet = "ใบแบ่ง"

// Branch is one receiving branch: its numeric code and Thai short name.
type Branch struct {
	Code string
	Name string
}

var poNumberRe = regexp.MustCompile(`\d{4,}`)

// ReadBranches opens a sample.xlsx-style workbook, reads the ใบแบ่ง sheet, and
// returns the purchase-order number (from A1) plus the ordered list of branches.
//
// In that sheet a *branch-header* row has an empty col A, a non-empty col B
// ("<code> <company name> ..."), and a non-empty col E (the Thai short name).
// Item rows have a sequence number in col A and are skipped.
func ReadBranches(path string) (po string, branches []Branch, err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	found := false
	for _, s := range f.GetSheetList() {
		if s == allocationSheet {
			found = true
			break
		}
	}
	if !found {
		return "", nil, fmt.Errorf("sheet %q not found (is this a sample-style file?)", allocationSheet)
	}

	rows, err := f.GetRows(allocationSheet)
	if err != nil {
		return "", nil, fmt.Errorf("read sheet %q: %w", allocationSheet, err)
	}
	if len(rows) > 0 {
		if m := poNumberRe.FindString(cell(rows[0], 0)); m != "" {
			po = m
		}
	}

	for _, row := range rows {
		a := strings.TrimSpace(cell(row, 0))
		b := strings.TrimSpace(cell(row, 1))
		e := strings.TrimSpace(cell(row, 4))
		if a == "" && b != "" && e != "" {
			branches = append(branches, Branch{Code: firstToken(b), Name: e})
		}
	}
	if len(branches) == 0 {
		return "", nil, fmt.Errorf("no branches found in sheet %q", allocationSheet)
	}
	return po, branches, nil
}

// cell returns the i-th cell of a row, or "" if the row is shorter (excelize trims
// trailing empty cells).
func cell(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}

// firstToken returns the first whitespace-delimited token (the branch code).
func firstToken(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
