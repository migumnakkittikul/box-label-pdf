package main

import _ "embed"

// The label font, baked into the binary so it runs standalone instead of needing
// a font file on disk. Sarabun, SIL Open Font License (see assets/OFL.txt).

//go:embed assets/Sarabun-Bold.ttf
var thaiFont []byte
