package main

import (
	"crypto/sha256"
	"strconv"

	"github.com/holiman/uint256"
)

// Generates a SHA256 hash of a string
func GenerateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return string(h.Sum(nil))
}

func GenerateHashOfList(lst []string) string {
	h := sha256.New()
	for _, s := range lst {
		h.Write([]byte(s))
	}
	return string(h.Sum(nil))
}

func CheckIfHashInHashRange(hash string, hr HashRange) bool {
	if hr.Low < hr.High {
		if hash > hr.Low && hash < hr.High {
			return true
		}
	} else {
		if hash > hr.Low || hash < hr.High {
			return true
		}
	}
	return false
}

// hashLow and hashHigh are both SHA256 hashes
func GetMidofHashes(hashLow string, hashHigh string) string {
	var l, h, mid uint256.Int
	l.SetBytes([]byte(hashLow))
	h.SetBytes([]byte(hashHigh))

	if hashLow < hashHigh {
		mid.Sub(&h, &l)
		mid.Div(&mid, uint256.NewInt(2))
		mid.Add(&mid, &l)
	} else {
		shaMax := GetSHAMax()
		mid.Sub(&shaMax, &l)
		mid.Add(&mid, &h)
		mid.Div(&mid, uint256.NewInt(2))
		var diff uint256.Int
		diff.Sub(&shaMax, &l)
		if mid.Cmp(&diff) == -1 {
			mid.Add(&mid, &l)
		} else {
			mid.Sub(&mid, &diff)
		}
	}

	return string(mid.Bytes())
}

func GetSHAMax() uint256.Int {
	var max uint256.Int
	b := make([]byte, 32)
	for i := 0; i < 32; i++ {
		b[i] = 255
	}
	max.SetBytes(b)
	return max
}

func GetMTLeafIndex(hash string, mt *MTNode) int {
	if !CheckIfHashInHashRange(hash, HashRange{mt.RngStart, mt.RngEnd}) {
		return -1
	}
	bin := calcMTLeafIndexBin(hash, mt)
	leafIndex, err := strconv.ParseInt(bin, 2, 64)
	if err != nil {
		panic(err)
	}
	return int(leafIndex)
}

func calcMTLeafIndexBin(hash string, mt *MTNode) string {
	if mt.IsLeaf {
		return ""
	}
	if CheckIfHashInHashRange(hash, HashRange{mt.Left.RngStart, mt.Left.RngEnd}) {
		return "0" + calcMTLeafIndexBin(hash, mt.Left)
	} else {
		return "1" + calcMTLeafIndexBin(hash, mt.Right)
	}
}
