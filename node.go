package main

import (
	"errors"
	"sync"
	"time"
)

// TODO: make a map of mutexes for each key

type Node struct {
	MList   *MemberList
	Config  *Config
	Info    *NodeInfo
	Engine  *Engine
	Router  *Router
	opsChan map[string]chan []byte
	mu      sync.Mutex
}

// TODO: functions for generate info, check if hash is in range

type NodeInfo struct {
	Name         string `json:"name"`
	Addr         string `json:"addr"`
	Port         string `json:"port"`
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
	n.Engine = &Engine{
		m: make(map[string]string),
	}
	n.Info.GetHash()
	n.opsChan = make(map[string]chan []byte)
	if config == nil {
		n.RequestConfig(seedNode)
	} else {
		n.Config = config
		n.Router = CreateRouter(config)
	}
	return &n
}

func (ni *NodeInfo) CheckIfHashInRange(hash string) bool {
	if ni.NodeHash > ni.PrevNodeHash {
		if hash > ni.PrevNodeHash && hash < ni.NodeHash {
			return true
		}
	} else {
		if hash > ni.PrevNodeHash || hash < ni.NodeHash {
			return true
		}
	}
	return false
}

// 8 character name, optionally padded with 0s
func (ni *NodeInfo) GetSenderName() string {
	return PadName(ni.Name)
}

// TODO: make this concurrent

func (n *Node) Read(key string) (value string, err error) {
	n.mu.Lock()
	n.opsChan[key] = make(chan []byte)
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
				if latestValue != DeletedHash {
					return latestValue, nil
				} else {
					return "", errors.New("key not found")
				}
			}
		case <-time.After(ReadTimeout):
			return "", errors.New("read timeout")
		}

	}
}

func (n *Node) Write(key string, value string) (err error) {
	n.mu.Lock()
	n.opsChan[key] = make(chan []byte)
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
			return errors.New("write timeout")
		}
	}
}

func (n *Node) Delete(key string) (err error) {
	return n.Write(key, DeletedHash)
}

const (
	ReadTimeout  = 3 * time.Second
	WriteTimeout = 3 * time.Second
)

/*
functions:
- DONE - read op
- DONE - write op
- DONE - coordinate with other nodes for replication
- DONE - if seed node, return config
- repair process
	- remap keys on node addition or removal
- API server
*/