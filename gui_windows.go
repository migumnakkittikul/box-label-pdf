//go:build windows

package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows"
)

var (
	modUser32   = windows.NewLazySystemDLL("user32.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procSystemParametersInfo = modUser32.NewProc("SystemParametersInfoW")
	procCreateWindowEx       = modUser32.NewProc("CreateWindowExW")
	procDestroyWindow        = modUser32.NewProc("DestroyWindow")
	procShowWindow           = modUser32.NewProc("ShowWindow")
	procSetForegroundWindow  = modUser32.NewProc("SetForegroundWindow")
	procGetModuleHandle      = modKernel32.NewProc("GetModuleHandleW")
)

// raiseDialogs makes this process the foreground process so the zenity dialogs open on
// top instead of behind the launching app. A -H windowsgui binary has no console and
// no window of its own, so zenity's own "bring my window to the front" logic finds
// nothing to raise and the dialog opens in the background (you have to click the
// taskbar). We fix it by (1) dropping the foreground-lock timeout so SetForegroundWindow
// is allowed, then (2) creating a tiny off-screen top-level window and making it the
// foreground window, which promotes our process to foreground and gives zenity a
// window to attach to. Returns a cleanup func that destroys the helper window.
func raiseDialogs() func() {
	const (
		spiSetForegroundLockTimeout = 0x2001
		spifSendChange              = 0x0002

		wsPopup        = 0x80000000
		wsExToolWindow = 0x00000080 // no taskbar button
		swShow         = 5
	)

	// Allow foreground changes for this session.
	procSystemParametersInfo.Call(spiSetForegroundLockTimeout, 0, 0, spifSendChange)

	hInst, _, _ := procGetModuleHandle.Call(0)
	class := syscall.StringToUTF16Ptr("STATIC")
	name := syscall.StringToUTF16Ptr("")
	hwnd, _, _ := procCreateWindowEx.Call(
		wsExToolWindow,
		uintptr(unsafe.Pointer(class)),
		uintptr(unsafe.Pointer(name)),
		wsPopup,
		0, 0, 1, 1, // 1x1 at the corner, effectively invisible
		0, 0, hInst, 0,
	)
	if hwnd != 0 {
		procShowWindow.Call(hwnd, swShow)
		procSetForegroundWindow.Call(hwnd)
	}
	return func() {
		if hwnd != 0 {
			procDestroyWindow.Call(hwnd)
		}
	}
}

// runGUI drives the Windows flow: pick file -> ask invoice -> pick save path ->
// generate -> report. Any cancelled dialog aborts cleanly.
func runGUI() {
	defer raiseDialogs()()

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
