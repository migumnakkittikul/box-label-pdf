package main

// genResult summarizes a successful generation, for the GUI/CLI to report.
type genResult struct {
	PO       string
	Branches int
	Labels   int
	Pages    int
}

// generate is the whole pipeline: read branches, build labels, render the PDF.
func generate(inPath, invoice, outPath string) (genResult, error) {
	po, branches, err := ReadBranches(inPath)
	if err != nil {
		return genResult{}, err
	}
	labels := BuildLabels(branches, invoice)
	if err := RenderPDF(labels, outPath, thaiFont); err != nil {
		return genResult{}, err
	}
	return genResult{
		PO:       po,
		Branches: len(branches),
		Labels:   len(labels),
		Pages:    PageCount(len(labels)),
	}, nil
}
