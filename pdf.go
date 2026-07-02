package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"

	"github.com/signintech/gopdf"
)

const mm2pt = 72.0 / 25.4

// Label geometry as fractions of the side length L: the header (logo) band, the
// key/colon/value column splits, and the body font size.
const (
	fHeader  = 0.156         // header (logos) band height
	fColKey  = 0.366         // right edge of the key column
	fDivider = 0.395         // key+colon | value boundary; header & footer divider x
	fFont    = 72.0 / 843.75 // font size as a fraction of L (~24pt at 100mm)
	nBodyRow = 6             // 5 info fields + footer
	medLine  = 1.0           // medium border line width (pt)
)

// Page layout: each page IS a 100x100mm sticker, one label per page.
const pageMM = 100.0 // sticker stock size (10x10cm)

// the five key labels, in order
var keyLabels = [5]string{"บริษัทผู้ส่ง", "รหัสบริษัท", "เลขที่ใบกำกับ", "รหัสสาขา", "สาขาผู้รับ"}

const (
	footerLeft  = "กล่องที่_______"
	footerRight = "จำนวนรวม__________"
)

type logoImg struct {
	holder gopdf.ImageHolder
	w, h   float64 // intrinsic pixel dimensions
}

func loadLogo(b []byte) (logoImg, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return logoImg{}, err
	}
	h, err := gopdf.ImageHolderByBytes(b)
	if err != nil {
		return logoImg{}, err
	}
	return logoImg{holder: h, w: float64(cfg.Width), h: float64(cfg.Height)}, nil
}

// RenderPDF writes the label PDF, one label per page, in order.
func RenderPDF(labels []Label, outPath string, fontData []byte) error {
	pdf, err := buildPDF(labels, fontData)
	if err != nil {
		return err
	}
	return pdf.WritePdf(outPath)
}

// buildPDF builds the document in memory so it can be inspected in tests.
func buildPDF(labels []Label, fontData []byte) (*gopdf.GoPdf, error) {
	pdf := &gopdf.GoPdf{}
	page := pageMM * mm2pt
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: page, H: page}})
	if err := pdf.AddTTFFontData("label", fontData); err != nil {
		return nil, fmt.Errorf("load font: %w", err)
	}
	left, err := loadLogo(logoLeft)
	if err != nil {
		return nil, fmt.Errorf("left logo: %w", err)
	}
	right, err := loadLogo(logoRight)
	if err != nil {
		return nil, fmt.Errorf("right logo: %w", err)
	}

	L := pageMM * mm2pt
	for _, lab := range labels {
		pdf.AddPage()
		drawLabel(pdf, 0, 0, L, lab, left, right)
	}
	return pdf, nil
}

// drawLabel draws one label with its upper-left corner at (ox, oy).
func drawLabel(pdf *gopdf.GoPdf, ox, oy, L float64, lab Label, left, right logoImg) {
	headerH := L * fHeader
	rowH := (L - headerH) / nBodyRow
	dividerX := L * fDivider
	keyW := L * fColKey
	colonW := dividerX - keyW
	valW := L - dividerX
	footerY := headerH + 5*rowH
	pad := L * 0.02
	valMaxW := valW - 2*pad
	fontSize := L * fFont

	pdf.SetStrokeColor(0, 0, 0)
	pdf.SetTextColor(0, 0, 0)

	// logos in the header band
	placeLogo(pdf, ox+pad, oy+pad, dividerX-2*pad, headerH-2*pad, left)
	placeLogo(pdf, ox+dividerX+pad, oy+pad, valW-2*pad, headerH-2*pad, right)

	// medium black lines
	pdf.SetLineWidth(medLine)
	pdf.RectFromUpperLeftWithStyle(ox, oy, L, L, "D")    // outer box
	pdf.Line(ox+dividerX, oy, ox+dividerX, oy+headerH)   // header divider (between logos)
	pdf.Line(ox, oy+headerH, ox+L, oy+headerH)           // under header
	pdf.Line(ox, oy+footerY, ox+L, oy+footerY)           // above footer
	pdf.Line(ox+dividerX, oy+footerY, ox+dividerX, oy+L) // footer divider

	// 5 info fields: key : value
	values := [5]string{lab.Sender, lab.CompanyCode, lab.Invoice, lab.BranchCode, lab.BranchName}
	for i := 0; i < 5; i++ {
		fy := oy + headerH + float64(i)*rowH
		pdf.SetFont("label", "", fontSize)
		drawCell(pdf, ox+pad, fy, keyW-pad, rowH, keyLabels[i], gopdf.Left|gopdf.Middle)
		drawCell(pdf, ox+keyW, fy, colonW, rowH, ":", gopdf.Center|gopdf.Middle)
		// shrink just this value if it would otherwise cross the right border
		pdf.SetFont("label", "", fitUniform(pdf, fontSize, []fitItem{{values[i], valMaxW}}))
		drawCell(pdf, ox+dividerX+pad, fy, valMaxW, rowH, values[i], gopdf.Left|gopdf.Middle)
	}

	// footer: กล่องที่___ | จำนวนรวม___
	fy := oy + footerY
	pdf.SetFont("label", "", fontSize)
	drawCell(pdf, ox+pad, fy, dividerX-2*pad, rowH, footerLeft, gopdf.Center|gopdf.Middle)
	drawCell(pdf, ox+dividerX+pad, fy, valMaxW, rowH, footerRight, gopdf.Center|gopdf.Middle)
}

type fitItem struct {
	text string
	maxW float64
}

// fitUniform returns the largest size <= base at which every item fits its maxW.
func fitUniform(pdf *gopdf.GoPdf, base float64, items []fitItem) float64 {
	for s := base; s > 6; s -= 0.5 {
		pdf.SetFont("label", "", s)
		ok := true
		for _, it := range items {
			w, err := pdf.MeasureTextWidth(it.text)
			if err == nil && w > it.maxW {
				ok = false
				break
			}
		}
		if ok {
			return s
		}
	}
	return 6
}

func placeLogo(pdf *gopdf.GoPdf, x, y, cw, ch float64, img logoImg) {
	s := math.Min(cw/img.w, ch/img.h)
	w, h := img.w*s, img.h*s
	pdf.ImageByHolder(img.holder, x+(cw-w)/2, y+(ch-h)/2, &gopdf.Rect{W: w, H: h})
}

func drawCell(pdf *gopdf.GoPdf, x, y, w, h float64, text string, align int) {
	pdf.SetXY(x, y)
	pdf.CellWithOption(&gopdf.Rect{W: w, H: h}, text, gopdf.CellOption{Align: align})
}
