package utils

import "math"

type AVLNode struct {
	Left   *AVLNode
	Right  *AVLNode
	Parent *AVLNode
	Height uint32
	Cnt    uint32
}

func avlHeight(Node *AVLNode) int {
	if Node == nil {
		return -1
	}
	if Node.Left == nil && Node.Right == nil {
		return 0
	}
	return int(Node.Height)
}

func avlUpdateHeight(Node *AVLNode) {
	Node.Height = uint32(1 + max(avlHeight(Node.Left), avlHeight(Node.Right)))
}

func avlBalanceFactor(Node *AVLNode) int {
	if Node == nil {
		return 0
	}
	return int(avlHeight(Node.Left)) - int(avlHeight(Node.Right))
}

func RRrotation(Root *AVLNode) *AVLNode {
	t1 := Root.Right
	t2 := t1.Left
	t1.Left = Root
	Root.Right = t2
	if t2 != nil {
		t2.Parent = Root
	}
	Root.Parent = t1
	t1.Parent = nil
	avlUpdateHeight(Root)
	avlUpdateHeight(t1)
	return t1
}

func LLrotation(Root *AVLNode) *AVLNode {
	t1 := Root.Left
	t2 := t1.Right
	t1.Right = Root
	Root.Left = t2
	if t2 != nil {
		t2.Parent = Root
	}
	Root.Parent = t1
	t1.Parent = nil
	avlUpdateHeight(Root)
	avlUpdateHeight(t1)
	return t1
}
