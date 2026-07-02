package main

import "testing"

// BenchmarkBuildPDF measures the in-memory document build for the sample workload.
func BenchmarkBuildPDF(b *testing.B) {
	_, branches, err := ReadBranches("sample.xlsx")
	if err != nil {
		b.Fatal(err)
	}
	labels := BuildLabels(branches, "IV-0001")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := buildPDF(labels, thaiFont); err != nil {
			b.Fatal(err)
		}
	}
}
