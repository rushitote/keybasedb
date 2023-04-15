package main

type MerkleTree struct {
	Root      *MTNode
	LeafNodes []*MTNode
}

// Merkle tree node
type MTNode struct {
	CurrDepth int
	Hash      string
	Left      *MTNode
	Right     *MTNode
	IsLeaf    bool
	RngStart  string // Includes this
	RngEnd    string // Excludes this
}

func CreateMerkleTree(hr HashRange) *MerkleTree {
	hashLow := hr.Low
	hashHigh := hr.High
	root := &MTNode{
		CurrDepth: 0,
		Hash:      "",
		Left:      nil,
		Right:     nil,
		IsLeaf:    false,
		RngStart:  hashLow,
		RngEnd:    hashHigh,
	}
	mt := &MerkleTree{
		Root: root,
	}
	mt.GenerateTree()
	return mt
}

func (mt *MerkleTree) GenerateTree() {
	mt.Root.CreateChildren(mt)
}

func (mtn *MTNode) CreateChildren(mt *MerkleTree) {
	if mtn.CurrDepth == MaxDepth {
		mtn.IsLeaf = true
		mt.LeafNodes = append(mt.LeafNodes, mtn)
		return
	}
	mid := GetMidofHashes(mtn.RngStart, mtn.RngEnd)
	mtn.Left = &MTNode{
		CurrDepth: mtn.CurrDepth + 1,
		Hash:      "",
		Left:      nil,
		Right:     nil,
		IsLeaf:    false,
		RngStart:  mtn.RngStart,
		RngEnd:    mid,
	}
	mtn.Right = &MTNode{
		CurrDepth: mtn.CurrDepth + 1,
		Hash:      "",
		Left:      nil,
		Right:     nil,
		IsLeaf:    false,
		RngStart:  mid,
		RngEnd:    mtn.RngEnd,
	}
	mtn.Left.CreateChildren(mt)
	mtn.Right.CreateChildren(mt)
}

const (
	MaxDepth int = 3
)
