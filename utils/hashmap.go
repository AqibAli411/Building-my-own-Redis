package utils

import (
	"fmt"
	"hash/fnv"
	"log"
)

type EntryType uint8
const (
	T_STR  EntryType = 1
	T_ZSET EntryType = 2
	T_ANY  EntryType = 3
	K_MAX_LOAD_FACTOR = 8
	K_RESIZING_WORK   = 128
)

type Entry struct {
	Node      HNode
	Key       string
	Type      EntryType
	Str       string
	HeapIndex int
}

type HNode struct {
	next  *HNode
	Hcode uint64
}

type HTable struct {
	Slots []*HNode
	Mask  uint64
	Size  uint64
}

type HMap struct {
	Newer      *HTable
	Older      *HTable
	MigratePos uint64
}

func HashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func htabInit(tab *HTable, n uint64) error {
	if (n & (n - 1)) != 0 {
		return fmt.Errorf("n must be a power of 2")
	}
	tab.Slots = make([]*HNode, n)
	tab.Mask = n - 1
	tab.Size = 0
	return nil
}

func htabInsert(table *HTable, node *HNode) {
	slot := node.Hcode & table.Mask
	node.next = table.Slots[slot]
	table.Slots[slot] = node
	table.Size++
}

func htabLookup(table *HTable, keyNode *HNode, eq func(*HNode, *HNode) bool) **HNode {
	if table == nil || len(table.Slots) == 0 {
		return nil
	}
	slot := keyNode.Hcode & table.Mask
	curr := &table.Slots[slot]
	for *curr != nil {
		if eq(*curr, keyNode) {
			return curr
		}
		curr = &(*curr).next
	}
	return nil
}

func htabDetach(htable *HTable, from **HNode) *HNode {
	if *from == nil {
		return nil
	}
	node := *from
	*from = node.next
	htable.Size--
	return node
}

func hmapStartResize(hmap *HMap) error {
	hmap.Older = hmap.Newer
	new_size := (1 + hmap.Newer.Mask) * 2
	hmap.Newer = &HTable{}
	err := htabInit(hmap.Newer, new_size)
	if err != nil {
		return err
	}
	hmap.MigratePos = 0
	return nil
}

func HmapMigrate(hmap *HMap) {
	if hmap.Older == nil {
		return
	}
	workDone := 0
	for workDone < K_RESIZING_WORK && hmap.Older.Size > 0 {
		if hmap.MigratePos >= hmap.Older.Mask+1 {
			break
		}
		from := &hmap.Older.Slots[hmap.MigratePos]
		if *from == nil {
			hmap.MigratePos++
			continue
		}
		node := htabDetach(hmap.Older, from)
		htabInsert(hmap.Newer, node)
		workDone++
	}
	if hmap.Older != nil && hmap.Older.Size == 0 {
		hmap.Older = nil
	}
}

func HmapLookup(hmap *HMap, key *HNode, eq func(*HNode, *HNode) bool) **HNode {
	HmapMigrate(hmap)
	result := htabLookup(hmap.Newer, key, eq)
	if result == nil {
		result = htabLookup(hmap.Older, key, eq)
	}
	return result
}

func HmapInsert(hmap *HMap, node *HNode) {
	if hmap.Newer == nil {
		hmap.Newer = &HTable{}
		htabInit(hmap.Newer, 4)
	}
	htabInsert(hmap.Newer, node)
	if hmap.Older == nil {
		if hmap.Newer.Size >= (hmap.Newer.Mask+1)*K_MAX_LOAD_FACTOR {
			hmapStartResize(hmap)
		}
	}
	HmapMigrate(hmap)
}

func HmapDetach(hmap *HMap, node *HNode, eq func(*HNode, *HNode) bool) *HNode {
	HmapMigrate(hmap)
	from := htabLookup(hmap.Newer, node, eq)
	if from != nil {
		return htabDetach(hmap.Newer, from)
	}
	from = htabLookup(hmap.Older, node, eq)
	if from != nil {
		return htabDetach(hmap.Older, from)
	}
	return nil
}

func HmapSize(hmap *HMap) uint64 {
	var size uint64
	if hmap.Newer != nil {
		size += hmap.Newer.Size
	}
	if hmap.Older != nil {
		size += hmap.Older.Size
	}
	return size
}
