package utils

import (
	"fmt"
	"unsafe"
)

type ZNode struct {
	Tree  AVLNode
	HNode HNode
	Score float64
	Name  string
}

type ZSet struct {
	Root *AVLNode
	HMap HMap
}
