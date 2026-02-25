package icon

import _ "embed"

//go:generate go run gen_icon.go

//go:embed icon.png
var Data []byte
