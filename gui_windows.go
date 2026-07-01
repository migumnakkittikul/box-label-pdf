//go:build windows

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
)

// runGUI drives the Windows flow: pick file -> ask invoice -> pick save path ->
// generate -> report. Any cancelled dialog aborts cleanly.
func runGUI() {
	inPath, err := zenity.SelectFile(
		zenity.Title("เลือกไฟล์ Excel (เช่น sample.xlsx)"),
		zenity.FileFilters{{Name: "Excel", Patterns: []string{"*.xlsx", "*.xls"}}},
	)
	if err != nil || inPath == "" {
		return
	}

	invoice, err := zenity.Entry(
		"กรอกเลขที่ใบกำกับ (เช่น IV-0001):",
		zenity.Title("เลขที่ใบกำกับ"),
	)
	if err != nil {
		return
	}
	invoice = strings.TrimSpace(invoice)
	if invoice == "" {
		zenity.Error("ยังไม่ได้กรอกเลขที่ใบกำกับ", zenity.Title("ผิดพลาด"))
		return
	}

	base := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
	outPath, err := zenity.SelectFileSave(
		zenity.Title("บันทึกไฟล์ PDF"),
		zenity.ConfirmOverwrite(),
		zenity.Filename(base+"_labels.pdf"),
		zenity.FileFilters{{Name: "PDF", Patterns: []string{"*.pdf"}}},
	)
	if err != nil || outPath == "" {
		return
	}
	if !strings.HasSuffix(strings.ToLower(outPath), ".pdf") {
		outPath += ".pdf"
	}

	res, err := generate(inPath, invoice, outPath)
	if err != nil {
		zenity.Error(err.Error(), zenity.Title("เกิดข้อผิดพลาด"))
		return
	}

	msg := fmt.Sprintf("เสร็จแล้ว!\n%d สาขา • %d ป้าย • %d หน้า (1 ป้าย/หน้า)\nไฟล์: %s\n\nพิมพ์ที่ขนาดจริง (100%%) ไม่ใช่ Fit to page",
		res.Branches, res.Labels, res.Pages, filepath.Base(outPath))
	zenity.Info(msg, zenity.Title("สำเร็จ"))
}
