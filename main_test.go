package main

import "testing"

func TestReadBranches(t *testing.T) {
	po, branches, err := ReadBranches("sample.xlsx")
	if err != nil {
		t.Fatalf("ReadBranches: %v", err)
	}
	if po != "1000000123" {
		t.Errorf("PO = %q, want 1000000123", po)
	}
	if got := len(branches); got != 8 {
		t.Fatalf("branch count = %d, want 8", got)
	}
	if branches[0].Code != "10010" || branches[0].Name != "สาขาตัวอย่าง 1" {
		t.Errorf("first branch = %+v, want {10010 สาขาตัวอย่าง 1}", branches[0])
	}
	last := branches[len(branches)-1]
	if last.Code != "10080" || last.Name != "สาขาตัวอย่าง 8" {
		t.Errorf("last branch = %+v, want {10080 สาขาตัวอย่าง 8}", last)
	}
	// No item rows should have leaked in: every code is a 5-digit number.
	for _, b := range branches {
		if len(b.Code) != 5 {
			t.Errorf("suspicious branch code %q (name %q)", b.Code, b.Name)
		}
	}
}

func TestBuildLabels(t *testing.T) {
	branches := []Branch{{"10010", "สาขาตัวอย่าง 1"}, {"10020", "สาขาตัวอย่าง 2"}}
	labels := BuildLabels(branches, "IV-0001")
	if len(labels) != 4 {
		t.Fatalf("label count = %d, want 4 (2 branches x2)", len(labels))
	}
	// Two consecutive copies per branch.
	if labels[0] != labels[1] {
		t.Errorf("labels[0] != labels[1]: %+v vs %+v", labels[0], labels[1])
	}
	if labels[0].BranchCode != "10010" || labels[2].BranchCode != "10020" {
		t.Errorf("branch ordering wrong: %q then %q", labels[0].BranchCode, labels[2].BranchCode)
	}
	l := labels[0]
	if l.Invoice != "IV-0001" || l.Sender != senderName || l.CompanyCode != companyCode {
		t.Errorf("constant fields wrong: %+v", l)
	}
	if l.BranchName != "สาขาตัวอย่าง 1" {
		t.Errorf("branch name = %q", l.BranchName)
	}
}

func TestPageCount(t *testing.T) {
	// One label per page: page count equals label count.
	cases := map[int]int{0: 0, 1: 1, 4: 4, 5: 5, 16: 16, 17: 17}
	for n, want := range cases {
		if got := PageCount(n); got != want {
			t.Errorf("PageCount(%d) = %d, want %d", n, got, want)
		}
	}
}

// TestEndToEndCounts ties the pieces together against the sample file.
func TestEndToEndCounts(t *testing.T) {
	_, branches, err := ReadBranches("sample.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	labels := BuildLabels(branches, "IV-0001")
	if len(labels) != 16 {
		t.Errorf("labels = %d, want 16", len(labels))
	}
	if PageCount(len(labels)) != 16 {
		t.Errorf("pages = %d, want 16 (one label per page)", PageCount(len(labels)))
	}
}
