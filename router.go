package main

import "sort"

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
	nodes := make([]*NodeInfo, 0)
	copy(nodes, r.cfg.Nodes)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].NodeHash < nodes[j].NodeHash
	})
	for i, node := range nodes {
		node.PrevNodeHash = nodes[(i-1)%len(nodes)].NodeHash
		node.NextNodeHash = nodes[(i+1)%len(nodes)].NodeHash
	}
}

func (r *Router) GetNodesInRange(key string, replicationFactor ReplicationFactor) []*NodeInfo {
	keyHash := GenerateHash(key)
	nodes := make([]*NodeInfo, 0)
	ringStartNode := -1
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

// TODO: Make this more efficient
