package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"aqib.builds/utils"
)

type ServerState struct {
	DB       utils.HMap
}

var Gdata ServerState

func NodeEq(a *utils.HNode, b *utils.HNode) bool {
	za := (*utils.Entry)(unsafe.Pointer(a))
	zb := (*utils.Entry)(unsafe.Pointer(b))
	return za.Key == zb.Key
}

func lookUpOrCreate(key string) *utils.Entry {
	entry := &utils.Entry{Key: key, Type: utils.T_ZSET, HeapIndex: -1}
	entry.Node.Hcode = utils.HashKey(key)
	from := utils.HmapLookup(&Gdata.DB, &entry.Node, NodeEq)
	if from != nil {
		return (*utils.Entry)(unsafe.Pointer(*from))
	}
	utils.HmapInsert(&Gdata.DB, &entry.Node)
	return entry
}

func lookUpExisting(key string, expectedType utils.EntryType) (*utils.Entry, error) {
	entry := &utils.Entry{Key: key, Type: expectedType, HeapIndex: -1}
	entry.Node.Hcode = utils.HashKey(key)
	node := utils.HmapLookup(&Gdata.DB, &entry.Node, NodeEq)
	if node == nil {
		return nil, nil
	}
	entry = (*utils.Entry)(unsafe.Pointer(*node))
	return entry, nil
}

func DoRequest(cmd []string, out *[]byte) {
	if len(cmd) == 0 {
		return
	}
	switch strings.ToLower(cmd[0]) {
	case "get":
		entry, _ := lookUpExisting(cmd[1], utils.T_STR)
		if entry == nil {
			*out = []byte("nil")
		} else {
			*out = []byte(entry.Str)
		}
	case "set":
		entry := lookUpOrCreate(cmd[1])
		entry.Str = cmd[2]
		entry.Type = utils.T_STR
		*out = []byte("OK")
	case "del":
		entry := &utils.Entry{Key: cmd[1]}
		entry.Node.Hcode = utils.HashKey(cmd[1])
		from := utils.HmapDetach(&Gdata.DB, &entry.Node, NodeEq)
		if from == nil {
			*out = []byte("0")
		} else {
			*out = []byte("1")
		}
	case "zadd":
		entry := lookUpOrCreate(cmd[1])
		score, _ := strconv.ParseFloat(cmd[2], 64)
		if entry.ZSet == nil {
			entry.ZSet = &utils.ZSet{}
		}
		utils.ZSetInsert(entry.ZSet, cmd[3], score)
		*out = []byte("OK")
	case "zscore":
		entry, _ := lookUpExisting(cmd[1], utils.T_ZSET)
		if entry == nil {
			*out = []byte("nil")
			return
		}
		found := utils.ZSetLookup(entry.ZSet, cmd[2])
		if found == nil {
			*out = []byte("nil")
		} else {
			*out = []byte(fmt.Sprintf("%f", found.Score))
		}
	}
}
