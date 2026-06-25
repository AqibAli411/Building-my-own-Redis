package utils

import "math"

type HeapItem struct {
	Val uint64
	Ref *int
}

func heapUp(heap []HeapItem, pos int) int {
	save := heap[pos]
	for pos > 0 {
		parentPos := int(math.Floor(float64(pos-1) / 2))
		if heap[parentPos].Val <= save.Val {
			break
		}
		heap[pos] = heap[parentPos]
		*heap[pos].Ref = pos
		pos = parentPos
	}
	heap[pos] = save
	*heap[pos].Ref = pos
	return pos
}
