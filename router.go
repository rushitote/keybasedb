package main

import (
	"sort"

	log "github.com/sirupsen/logrus"
)

// Router is for routing requests to other nodes
type Router struct {
	cfg *Config
}

// CreateRouter creates a new router and a virtual ring of nodes
func CreateRouter(cfg *Config) *Router {
	r := &Router{cfg: cfg}
	r.CreateRing()
	return r
}

func (r *Router) CreateRing() {
	log.Infof("Creating ring of nodes in the cluster")
	for _, node := range r.cfg.Nodes {
		node.GetHash()
	}
	nodes := make([]*NodeInfo, len(r.cfg.Nodes))
	copy(nodes, r.cfg.Nodes)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].NodeHash < nodes[j].NodeHash
	})
	for i, node := range nodes {
		node.PrevNodeHash = nodes[(i-1+len(nodes))%len(nodes)].NodeHash
		node.NextNodeHash = nodes[(i+1)%len(nodes)].NodeHash
	}
}

func (r *Router) GetNodesInRange(key string, replicationFactor ReplicationFactor) []*NodeInfo {
	keyHash := GenerateHash(key)
	nodes := make([]*NodeInfo, 0)
	ringStartNode := -1
	// TODO: Make this more efficient
	for i, nodeInfo := range r.cfg.Nodes {
		if nodeInfo.CheckIfHashInRange(keyHash) {
			ringStartNode = i
			break
		}
	}
	for i := 0; i < int(replicationFactor); i++ {
		nodes = append(nodes, r.cfg.Nodes[(ringStartNode+i)%len(r.cfg.Nodes)])
	}
	return nodes
}

func (r *Router) GetHashRangesForRepair(currNode, otherNode string) []HashRange {
	currNodeHash := GenerateHash(currNode)
	otherNodeHash := GenerateHash(otherNode)
	hashRanges := make([]HashRange, 0)
	for i, node := range r.cfg.Nodes {
		isRangePresent := false
		for j := i; j < i+int(r.cfg.ReplicationFactor); j++ {
			if r.cfg.Nodes[j%len(r.cfg.Nodes)].NodeHash == otherNodeHash {
				isRangePresent = true
				break
			}
		}
		if !isRangePresent {
			continue
		}
		isRangePresent = false
		for j := i; j < i+int(r.cfg.ReplicationFactor); j++ {
			if r.cfg.Nodes[j%len(r.cfg.Nodes)].NodeHash == currNodeHash {
				isRangePresent = true
				break
			}
		}
		if isRangePresent {
			hashRanges = append(hashRanges, HashRange{Low: node.PrevNodeHash, High: node.NodeHash})
		}
	}
	return hashRanges
}
