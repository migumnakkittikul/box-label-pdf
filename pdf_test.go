package main

import (
	"testing"

	"github.com/signintech/gopdf"
)

func newTestPDF(t *testing.T) *gopdf.GoPdf {
	t.Helper()
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := pdf.AddTTFFontData("label", thaiFont); err != nil {
		t.Fatal(err)
	}
	return pdf
}

func TestFitUniform(t *testing.T) {
	pdf := newTestPDF(t)
	const base = 24.0

	// Plenty of room -> keeps the base size.
	if got := fitUniform(pdf, base, []fitItem{{"10010", 500}}); got != base {
		t.Errorf("short text in wide column = %.1f, want %.1f", got, base)
	}
	// Impossible width -> shrinks below base but not below the 6pt floor.
	got := fitUniform(pdf, base, []fitItem{{"เลขที่ใบกำกับยาวๆ", 20}})
	if got >= base || got < 6 {
		t.Errorf("long text in narrow column = %.1f, want in [6, %.1f)", got, base)
	}
	// Narrower column never yields a larger size (monotonic).
	wide := fitUniform(pdf, base, []fitItem{{"บริษัทผู้ส่ง", 200}})
	narrow := fitUniform(pdf, base, []fitItem{{"บริษัทผู้ส่ง", 60}})
	if narrow > wide {
		t.Errorf("narrower column gave larger size: narrow=%.1f wide=%.1f", narrow, wide)
	}
}

func TestRenderLongBranchNameNoPanic(t *testing.T) {
	labels := []Label{{senderName, companyCode, "IV-0001", "10000",
		"สาขาที่มีชื่อยาวมากเกินปกติเพื่อทดสอบการย่อขนาดตัวอักษรให้พอดีกับคอลัมน์"}}
	if _, err := buildPDF(labels, thaiFont); err != nil {
		t.Fatalf("long branch name should still render: %v", err)
	}
}
