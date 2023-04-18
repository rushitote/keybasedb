package main

import (
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Node struct {
	MList    *MemberList
	Config   *Config
	Info     *NodeInfo
	Engine   *Engine
	Router   *Router
	Server   *APIServer
	Graph    *Graph
	opsChan  map[string]chan []byte
	opsMutex map[string]*sync.RWMutex
	mu       sync.Mutex
}

type NodeInfo struct {
	Name         string `json:"name"`
	Addr         string `json:"addr"`
	Port         string `json:"port"`
	APIPort      string `json:"api_port"`
	NodeHash     string `json:"node_hash"`
	PrevNodeHash string `json:"prev_node_hash"`
	NextNodeHash string `json:"next_node_hash"`
}

func (ni *NodeInfo) GetHash() string {
	if ni.NodeHash == "" {
		ni.NodeHash = GenerateHash(ni.Name)
	}
	return ni.NodeHash
}

func StartNode(config *Config, currNode *NodeInfo, seedNode *NodeInfo) *Node {
	var n Node
	n.MList = CreateMemberList(currNode, seedNode, n.ProcessMsg)
	n.Info = currNode
	n.Engine = CreateEngine(n.Info.Name)
	n.Info.GetHash()
	n.Server = InitServer(&n, n.Read, n.Write, n.Delete, n.Repair)
	n.opsChan = make(map[string]chan []byte)
	n.opsMutex = make(map[string]*sync.RWMutex)
	n.Graph = &Graph{}
	if config == nil {
		n.RequestConfigRep(seedNode)
	} else {
		n.Config = config
		n.Router = CreateRouter(config)
		n.Config.State = STABLE
	}
	log.Infof("Node %s started", n.Info.Name)
	return &n
}

func (n *Node) RequestConfigRep(seedNode *NodeInfo) {
	n.RequestConfig(seedNode)
	time.Sleep(1 * time.Second)
	if n.Config == nil {
		n.RequestConfigRep(seedNode)
	}
}

func (ni *NodeInfo) CheckIfHashInRange(hash string) bool {
	return CheckIfHashInHashRange(hash, HashRange{Low: ni.PrevNodeHash, High: ni.NodeHash})
}

// 8 character name, optionally padded with 0s
func (ni *NodeInfo) GetSenderName() string {
	return PadName(ni.Name)
}

// TODO: make this concurrent

func (n *Node) Read(key string) (value string, err error) {
	log.Infof("Read request for key=%s", key)
	if n.Config.State != STABLE {
		return "", errors.New(CLUSTER_NOT_STABLE)
	}

	m, ok := n.opsMutex[key]
	if !ok {
		m = &sync.RWMutex{}
		n.opsMutex[key] = m
	}
	m.RLock()
	defer m.RUnlock()
	n.mu.Lock()
	n.opsChan[key] = make(chan []byte, n.Config.ReplicationFactor)
	n.mu.Unlock()

	readNum := 0
	nodesWithKey := n.Router.GetNodesInRange(key, n.Config.ReplicationFactor)
	for _, node := range nodesWithKey {
		n.RequestRead(key, node.Name)
	}
	var latestValue, lastTimestamp string
	for {
		select {
		case msg := <-n.opsChan[key]:
			readNum++
			ts := GetTimestampFromValue(string(msg))
			if ts != "" && (lastTimestamp == "" || lastTimestamp < ts) {
				latestValue = GetValueTextFromValue(string(msg))
				lastTimestamp = ts
			}
			if readNum >= n.Config.MinReadsRequired {
				if latestValue != DeletedHash && latestValue != "" {
					return latestValue, nil
				} else {
					return "", errors.New(KEY_NOT_FOUND)
				}
			}
		case <-time.After(ReadTimeout):
			return "", errors.New(READ_TIMEOUT + " for key = " + key)
		}

	}
}

func (n *Node) Write(key string, value string) (err error) {
	log.Infof("Write request for key=%s", key)
	if n.Config.State != STABLE {
		return errors.New(CLUSTER_NOT_STABLE)
	}

	m, ok := n.opsMutex[key]
	if !ok {
		m = &sync.RWMutex{}
		n.opsMutex[key] = m
	}
	m.Lock()
	defer m.Unlock()
	n.mu.Lock()
	n.opsChan[key] = make(chan []byte, n.Config.ReplicationFactor)
	n.mu.Unlock()

	writeNum := 0
	value = AddTimestampToValue(value)
	nodesWithKey := n.Router.GetNodesInRange(key, n.Config.ReplicationFactor)
	for _, node := range nodesWithKey {
		n.RequestWrite(key, value, node.Name)
	}
	for {
		select {
		case <-n.opsChan[key]:
			writeNum++
			if writeNum >= n.Config.MinWritesRequired {
				return nil
			}
		case <-time.After(WriteTimeout):
			return errors.New(CLUSTER_NOT_STABLE)
		}
	}
}

func (n *Node) Delete(key string) (err error) {
	return n.Write(key, DeletedHash)
}

func (n *Node) Repair(otherNode string) (err error) {
	log.Infof("Repair request for node=%s", otherNode)
	if n.Config.State != STABLE {
		return errors.New(CLUSTER_NOT_STABLE)
	}
	n.Config.State = UNSTABLE
	hashRanges := n.Router.GetHashRangesForRepair(n.Info.Name, otherNode)
	for _, hashRange := range hashRanges {
		err := n.RepairHashRange(otherNode, hashRange)
		if err != nil {
			return err
		}
	}
	n.Config.State = STABLE
	return nil
}

func (n *Node) RepairHashRange(otherNode string, hashRange HashRange) (err error) {
	n.Engine.Stream(func(key, value string) error {
		if CheckIfHashInHashRange(key, hashRange) {
			n.RequestRepair(key, value, otherNode)
		}
		return nil
	})
	return nil
}

const (
	ReadTimeout  = 10 * time.Second
	WriteTimeout = 10 * time.Second
)

/*
functions:
- DONE - read op
- DONE - write op
- DONE - coordinate with other nodes for replication
- DONE - if seed node, return config
- DONE - integrate badger
- repair process (merkle tree)
	- remap keys on node addition or removal
- DONE - API server
- Remove panics
- Add two phase commit
*/
