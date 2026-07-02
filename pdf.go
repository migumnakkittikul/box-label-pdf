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

// RenderPDF writes the label PDF: one 100x100mm label per page, in order.
func RenderPDF(labels []Label, outPath string, fontData []byte) error {
	pdf, err := buildPDF(labels, fontData)
	if err != nil {
		return err
	}
	return pdf.WritePdf(outPath)
}

// buildPDF builds the document in memory (no file I/O), so it can be benchmarked
// and inspected in tests.
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
	lay := newLabelLayout(pdf, L)
	for _, lab := range labels {
		pdf.AddPage()
		lay.draw(pdf, 0, 0, lab, left, right)
	}
	return pdf, nil
}

// the five key labels, in order (constant for every label)
var keyLabels = [5]string{"บริษัทผู้ส่ง", "รหัสบริษัท", "เลขที่ใบกำกับ", "รหัสสาขา", "สาขาผู้รับ"}

const (
	footerLeft  = "กล่องที่_______"
	footerRight = "จำนวนรวม__________"
)

// labelLayout holds the geometry and font sizes that are identical for every label,
// computed once. The per-value fit size is memoized so each distinct value (the
// constant sender line, each branch name) is measured only once across the whole run.
type labelLayout struct {
	pdf                          *gopdf.GoPdf
	L, headerH, rowH             float64
	dividerX, keyW, colonW, valW float64
	footerY, pad, valMaxW        float64
	bodySize                     float64 // uniform key size
	flSize, frSize               float64 // footer sizes
	valCache                     map[string]float64
}

func newLabelLayout(pdf *gopdf.GoPdf, L float64) *labelLayout {
	l := &labelLayout{pdf: pdf, L: L, valCache: map[string]float64{}}
	l.headerH = L * fHeader
	l.rowH = (L - l.headerH) / nBodyRow
	l.dividerX = L * fDivider
	l.keyW = L * fColKey
	l.colonW = l.dividerX - l.keyW
	l.valW = L - l.dividerX
	l.footerY = l.headerH + 5*l.rowH
	l.pad = L * 0.02
	l.valMaxW = l.valW - 2*l.pad
	fontSize := L * fFont

	// Keys share one size, shrunk just enough that the longest key fits its column.
	// A narrow font yields a larger size, a wider font a smaller one.
	keyItems := make([]fitItem, len(keyLabels))
	for i, k := range keyLabels {
		keyItems[i] = fitItem{k, l.keyW - l.pad}
	}
	l.bodySize = fitUniform(pdf, fontSize, keyItems)
	l.flSize = fitUniform(pdf, fontSize, []fitItem{{footerLeft, l.dividerX - 2*l.pad}})
	l.frSize = fitUniform(pdf, fontSize, []fitItem{{footerRight, l.valMaxW}})
	return l
}

// valSize returns the (memoized) font size for a value: bodySize, shrunk on its own
// only if the value would otherwise cross the right border.
func (l *labelLayout) valSize(text string) float64 {
	if v, ok := l.valCache[text]; ok {
		return v
	}
	v := fitUniform(l.pdf, l.bodySize, []fitItem{{text, l.valMaxW}})
	l.valCache[text] = v
	return v
}

func (l *labelLayout) draw(pdf *gopdf.GoPdf, ox, oy float64, lab Label, left, right logoImg) {
	pdf.SetStrokeColor(0, 0, 0)
	pdf.SetTextColor(0, 0, 0)

	// logos in the header band
	placeLogo(pdf, ox+l.pad, oy+l.pad, l.dividerX-2*l.pad, l.headerH-2*l.pad, left)
	placeLogo(pdf, ox+l.dividerX+l.pad, oy+l.pad, l.valW-2*l.pad, l.headerH-2*l.pad, right)

	// medium black lines
	pdf.SetLineWidth(medLine)
	pdf.RectFromUpperLeftWithStyle(ox, oy, l.L, l.L, "D")        // outer box
	pdf.Line(ox+l.dividerX, oy, ox+l.dividerX, oy+l.headerH)     // header divider (between logos)
	pdf.Line(ox, oy+l.headerH, ox+l.L, oy+l.headerH)             // under header
	pdf.Line(ox, oy+l.footerY, ox+l.L, oy+l.footerY)             // above footer
	pdf.Line(ox+l.dividerX, oy+l.footerY, ox+l.dividerX, oy+l.L) // footer divider

	// 5 info fields: key : value
	values := [5]string{lab.Sender, lab.CompanyCode, lab.Invoice, lab.BranchCode, lab.BranchName}
	for i := 0; i < 5; i++ {
		fy := oy + l.headerH + float64(i)*l.rowH
		pdf.SetFont("label", "", l.bodySize)
		drawCell(pdf, ox+l.pad, fy, l.keyW-l.pad, l.rowH, keyLabels[i], gopdf.Left|gopdf.Middle)
		drawCell(pdf, ox+l.keyW, fy, l.colonW, l.rowH, ":", gopdf.Center|gopdf.Middle)
		pdf.SetFont("label", "", l.valSize(values[i]))
		drawCell(pdf, ox+l.dividerX+l.pad, fy, l.valMaxW, l.rowH, values[i], gopdf.Left|gopdf.Middle)
	}

	// footer: กล่องที่___ | จำนวนรวม___
	fy := oy + l.footerY
	pdf.SetFont("label", "", l.flSize)
	drawCell(pdf, ox+l.pad, fy, l.dividerX-2*l.pad, l.rowH, footerLeft, gopdf.Center|gopdf.Middle)
	pdf.SetFont("label", "", l.frSize)
	drawCell(pdf, ox+l.dividerX+l.pad, fy, l.valMaxW, l.rowH, footerRight, gopdf.Center|gopdf.Middle)
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
