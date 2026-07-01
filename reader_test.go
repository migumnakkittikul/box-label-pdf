package main

import (
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// makeXlsx writes a temp .xlsx with the given sheet name and rows (only non-empty
// cells are set; empty strings are left blank so leading/empty columns are exercised).
func makeXlsx(t *testing.T, sheet string, rows [][]string) string {
	t.Helper()
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheet)
	for r, row := range rows {
		for c, v := range row {
			if v == "" {
				continue
			}
			cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
			if err := f.SetCellStr(sheet, cell, v); err != nil {
				t.Fatal(err)
			}
		}
	}
	p := filepath.Join(t.TempDir(), "test.xlsx")
	if err := f.SaveAs(p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReaderSyntheticBasic(t *testing.T) {
	rows := [][]string{
		{"ใบสั่งซื้อเลขที่ 1000000123"},
		{"ลำดับ", "SKU No.", "รุ่น", "จำนวน", "สาขา"},
		{"", "10010 Example Retail Co. (Branch 1)", "", "", "สาขาตัวอย่าง 1"},
		{"1", "70010001", "M-100", "3", ""},
		{"2", "70010002", "M-200", "5", ""},
		{"", "10020 Example Retail Co. (Branch 2)", "", "", "สาขาตัวอย่าง 2"},
		{"1", "70010003", "M-300", "1", ""},
	}
	po, branches, err := ReadBranches(makeXlsx(t, allocationSheet, rows))
	if err != nil {
		t.Fatal(err)
	}
	if po != "1000000123" {
		t.Errorf("po = %q, want 1000000123", po)
	}
	if len(branches) != 2 {
		t.Fatalf("branches = %d, want 2 (item rows must be skipped)", len(branches))
	}
	if branches[0] != (Branch{"10010", "สาขาตัวอย่าง 1"}) {
		t.Errorf("branch0 = %+v", branches[0])
	}
	if branches[1] != (Branch{"10020", "สาขาตัวอย่าง 2"}) {
		t.Errorf("branch1 = %+v", branches[1])
	}
}

func TestReaderTrimsAndExtractsCode(t *testing.T) {
	rows := [][]string{
		{"ใบสั่งซื้อเลขที่ 1"},
		{"ลำดับ", "SKU No.", "", "", "สาขา"},
		{"", "  10030   Example Retail Co. (Branch 3)  ", "", "", "  สาขาตัวอย่าง 3  "},
	}
	_, branches, err := ReadBranches(makeXlsx(t, allocationSheet, rows))
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 1 || branches[0] != (Branch{"10030", "สาขาตัวอย่าง 3"}) {
		t.Errorf("got %+v, want one {10030 สาขาตัวอย่าง 3}", branches)
	}
}

func TestReaderSkipsNonBranchRows(t *testing.T) {
	rows := [][]string{
		{"ใบสั่งซื้อเลขที่ 7"},
		{"ลำดับ", "SKU No.", "", "", "สาขา"},                                // header: col A non-empty -> skip
		{"5", "x", "", "", "ไม่ใช่สาขา"},                                    // item-like: col A set -> skip even with E set
		{"", "10090 only B set", "", "", ""},                                // col E empty -> not a branch header -> skip
		{"", "", "", "", "มีแต่ชื่อ"},                                       // col B empty -> skip
		{"", "10100 Example Retail Co. (Branch 5)", "", "", "สาขาตัวอย่าง"}, // the only branch header
	}
	_, branches, err := ReadBranches(makeXlsx(t, allocationSheet, rows))
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 1 || branches[0] != (Branch{"10100", "สาขาตัวอย่าง"}) {
		t.Errorf("got %+v, want only {10100 สาขาตัวอย่าง}", branches)
	}
}

func TestReaderMissingSheet(t *testing.T) {
	rows := [][]string{{"", "10010 X", "", "", "สาขาตัวอย่าง"}}
	_, _, err := ReadBranches(makeXlsx(t, "WrongSheet", rows))
	if err == nil {
		t.Fatal("expected error for missing ใบแบ่ง sheet")
	}
}

func TestReaderNoBranches(t *testing.T) {
	rows := [][]string{
		{"ใบสั่งซื้อเลขที่ 1"},
		{"ลำดับ", "SKU No.", "", "", "สาขา"},
	}
	_, _, err := ReadBranches(makeXlsx(t, allocationSheet, rows))
	if err == nil {
		t.Fatal("expected error when no branches present")
	}
}

func TestReaderPONotFound(t *testing.T) {
	rows := [][]string{
		{"ใบสั่งซื้อ (ไม่มีเลข)"},
		{"ลำดับ", "SKU No.", "", "", "สาขา"},
		{"", "10010 X", "", "", "สาขาตัวอย่าง"},
	}
	po, _, err := ReadBranches(makeXlsx(t, allocationSheet, rows))
	if err != nil {
		t.Fatal(err)
	}
	if po != "" {
		t.Errorf("po = %q, want empty when no number in A1", po)
	}
}

func TestReaderMissingFile(t *testing.T) {
	if _, _, err := ReadBranches("does-not-exist.xlsx"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFirstToken(t *testing.T) {
	cases := map[string]string{
		"":                         "",
		"   ":                      "",
		"10010":                    "10010",
		"10010 Example Retail Co.": "10010",
		"  10030   Example  ":      "10030",
	}
	for in, want := range cases {
		if got := firstToken(in); got != want {
			t.Errorf("firstToken(%q) = %q, want %q", in, got, want)
		}
	}
}

// Sample-file cross-checks beyond the basic count test.
func TestReaderRealFileIntegrity(t *testing.T) {
	_, branches, err := ReadBranches("sample.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, b := range branches {
		if len(b.Code) != 5 {
			t.Errorf("non-5-digit code %q (%q)", b.Code, b.Name)
		}
		for _, r := range b.Code {
			if r < '0' || r > '9' {
				t.Errorf("non-numeric code %q", b.Code)
				break
			}
		}
		if b.Name == "" {
			t.Errorf("empty name for %q", b.Code)
		}
		seen[b.Code] = true
	}
	if len(seen) != 8 {
		t.Errorf("distinct codes = %d, want 8", len(seen))
	}
}
