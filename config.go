package main

import "encoding/json"

// Config is configuration shared by all nodes in the cluster
type Config struct {
	ReplicationFactor ReplicationFactor `json:"replication_factor"`
	ConsistencyLevel  ConsistencyLevel  `json:"consistency_level"`
	MinReadsRequired  int               `json:"min_reads_required"`
	MinWritesRequired int               `json:"min_writes_required"`
	Nodes             []*NodeInfo       `json:"nodes"`
	State             ClusterState      `json:"state"`
}

type ReplicationFactor int

const (
	ONE   ReplicationFactor = 1
	TWO   ReplicationFactor = 2
	THREE ReplicationFactor = 3
)

type ConsistencyLevel int

const (
	QUORUM ConsistencyLevel = iota
	ALL
)

type ClusterState string

const (
	STABLE   ClusterState = "STABLE"   // General state
	UNSTABLE ClusterState = "UNSTABLE" // Unstable state
)

func CreateConfig(replicationFactor ReplicationFactor, consistencyLevel ConsistencyLevel, nodes []*NodeInfo) *Config {
	var minReadsRequired, minWritesRequired int

	if len(nodes) < minReadsRequired {
		panic("Not enough nodes to satisfy read quorum")
	}

	if replicationFactor <= 0 {
		panic("Replication factor must be greater than 0")
	}

	if consistencyLevel == QUORUM {
		minReadsRequired = int(replicationFactor)/2 + 1
		minWritesRequired = int(replicationFactor)/2 + 1
	} else if consistencyLevel == ALL {
		minReadsRequired = int(replicationFactor)
		minWritesRequired = int(replicationFactor)
	} else {
		panic("Invalid consistency level")
	}

	return &Config{
		ReplicationFactor: replicationFactor,
		ConsistencyLevel:  consistencyLevel,
		MinReadsRequired:  minReadsRequired,
		MinWritesRequired: minWritesRequired,
		Nodes:             nodes,
		State:             UNSTABLE,
	}
}

func (c *Config) SerializeConfig() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return b
}

func DeserializeConfig(b []byte) *Config {
	var c *Config
	err := json.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}
	return c
}
