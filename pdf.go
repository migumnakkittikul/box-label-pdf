package main

import (
	"fmt"

	"github.com/signintech/gopdf"
)

const mm2pt = 72.0 / 25.4

// Label geometry as fractions of the side length L: the header band, the
// key/colon/value column splits, and the body font size.
const (
	fHeader  = 0.156         // header band height
	fColKey  = 0.366         // right edge of the key column
	fDivider = 0.395         // key+colon | value boundary
	fFont    = 72.0 / 843.75 // font size as a fraction of L (~24pt at 100mm)
	nBodyRow = 6             // 5 info fields + footer
)

// fontPath is the label font, loaded from disk for now.
const fontPath = "assets/Sarabun-Bold.ttf"

// the five key labels, in order
var keyLabels = [5]string{"บริษัทผู้ส่ง", "รหัสบริษัท", "เลขที่ใบกำกับ", "รหัสสาขา", "สาขาผู้รับ"}

const (
	footerLeft  = "กล่องที่_______"
	footerRight = "จำนวนรวม__________"
)

// RenderPDF writes the label PDF, one label per page, in order.
func RenderPDF(labels []Label, outPath string) error {
	pdf, err := buildPDF(labels)
	if err != nil {
		return err
	}
	return pdf.WritePdf(outPath)
}

// buildPDF builds the document in memory so it can be inspected in tests.
func buildPDF(labels []Label) (*gopdf.GoPdf, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := pdf.AddTTFFont("label", fontPath); err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}
	L := 500.0
	for _, lab := range labels {
		pdf.AddPage()
		drawLabel(pdf, 0, 0, L, lab)
	}
	return pdf, nil
}

// drawLabel draws one label with its upper-left corner at (ox, oy).
func drawLabel(pdf *gopdf.GoPdf, ox, oy, L float64, lab Label) {
	headerH := L * fHeader
	rowH := (L - headerH) / nBodyRow
	dividerX := L * fDivider
	keyW := L * fColKey
	colonW := dividerX - keyW
	valW := L - dividerX
	footerY := headerH + 5*rowH
	pad := L * 0.02
	fontSize := L * fFont

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("label", "", fontSize)

	values := [5]string{lab.Sender, lab.CompanyCode, lab.Invoice, lab.BranchCode, lab.BranchName}
	for i := 0; i < 5; i++ {
		fy := oy + headerH + float64(i)*rowH
		drawCell(pdf, ox+pad, fy, keyW-pad, rowH, keyLabels[i], gopdf.Left|gopdf.Middle)
		drawCell(pdf, ox+keyW, fy, colonW, rowH, ":", gopdf.Center|gopdf.Middle)
		drawCell(pdf, ox+dividerX+pad, fy, valW-pad, rowH, values[i], gopdf.Left|gopdf.Middle)
	}

	// footer line
	fy := oy + footerY
	drawCell(pdf, ox+pad, fy, valW, rowH, footerLeft+"  "+footerRight, gopdf.Left|gopdf.Middle)
}

func drawCell(pdf *gopdf.GoPdf, x, y, w, h float64, text string, align int) {
	pdf.SetXY(x, y)
	pdf.CellWithOption(&gopdf.Rect{W: w, H: h}, text, gopdf.CellOption{Align: align})
}
