package main

// Constants for the sender, matching the label layout. Only the invoice number
// and the branch fields vary; these two are fixed.
const (
	senderName  = "บจก. ตัวอย่างซัพพลาย"
	companyCode = "100200"

	// labelsPerBranch is how many copies of each branch's label to print.
	labelsPerBranch = 2

	// labelsPerPage is how many labels go on each page (one 100x100mm label per page).
	labelsPerPage = 1
)

// Label is the content of one box-front label. Field order matches the label layout:
// บริษัทผู้ส่ง / รหัสบริษัท / เลขที่ใบกำกับ / รหัสสาขา / สาขาผู้รับ.
type Label struct {
	Sender      string
	CompanyCode string
	Invoice     string
	BranchCode  string
	BranchName  string
}

// BuildLabels expands the branch list into labels, emitting labelsPerBranch copies
// of each branch consecutively, all carrying the given invoice number.
func BuildLabels(branches []Branch, invoice string) []Label {
	out := make([]Label, 0, len(branches)*labelsPerBranch)
	for _, b := range branches {
		for i := 0; i < labelsPerBranch; i++ {
			out = append(out, Label{
				Sender:      senderName,
				CompanyCode: companyCode,
				Invoice:     invoice,
				BranchCode:  b.Code,
				BranchName:  b.Name,
			})
		}
	}
	return out
}

// PageCount returns how many pages n labels need at labelsPerPage per page.
func PageCount(n int) int {
	return (n + labelsPerPage - 1) / labelsPerPage
}
