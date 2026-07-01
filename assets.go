package main

import _ "embed"

// Header logos and the label font, baked into the binary so it runs standalone.
// The logos are placeholders - drop in your own PNGs to rebrand. The font is
// Sarabun (SIL Open Font License, see assets/OFL.txt).

//go:embed assets/logo_left.png
var logoLeft []byte

//go:embed assets/logo_right.png
var logoRight []byte

//go:embed assets/Sarabun-Bold.ttf
var thaiFont []byte
