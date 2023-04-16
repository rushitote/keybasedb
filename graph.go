package main

import (
	"encoding/json"
	"math"
)

type Graph struct {
	Nodes  map[string]*NeighbourDeegrees
	NDChan map[string]chan []byte
	MaxDeg int
	n      *Node
}

type NeighbourDeegrees struct {
	Key        string         `json:"key"`
	Neighbours map[string]int `json:"neighbours"`
}

func (g *Graph) ReconstructGraph(n *Node) {
	g.Nodes = make(map[string]*NeighbourDeegrees)
	g.NDChan = make(map[string]chan []byte)
	g.n = n
	g.MaxDeg = 2
	n.Engine.Stream(func(key string, value string) error {
		g.Nodes[key] = ReconstructNode(key, n, g.MaxDeg)
		return nil
	})
}

func ReconstructNode(key string, n *Node, maxDegree int) *NeighbourDeegrees {
	bfsQueue := make([]string, 0)
	bfsQueue = append(bfsQueue, key)

	visited := make(map[string]bool)
	neighbourDegrees := make(map[string]int)

	currDist := 0
	for len(bfsQueue) > 0 {
		sz := len(bfsQueue)
		for i := 0; i < sz; i++ {
			currVertex := bfsQueue[0]
			bfsQueue = bfsQueue[1:]
			visited[currVertex] = true
			neighbourDegrees[currVertex] = currDist

			neighbours, err := n.Read(currVertex)
			if err != nil && err.Error() != KEY_NOT_FOUND {
				panic(err)
			} else if err != nil && err.Error() == KEY_NOT_FOUND {
				continue
			}
			var vn VertexNeighbours
			err = json.Unmarshal([]byte(neighbours), &vn)
			if err != nil {
				panic(err)
			}
			for _, n := range vn.Neighbours {
				if !visited[n] {
					bfsQueue = append(bfsQueue, n)
				}
			}
		}
		currDist++
		if currDist > maxDegree {
			break
		}
	}

	return &NeighbourDeegrees{Neighbours: neighbourDegrees}
}

func (g *Graph) FindDegreeBetween(v1 string, v2 string) int {
	n1 := g.n.Router.GetNodesInRange(v1, 1)[0]
	n2 := g.n.Router.GetNodesInRange(v2, 1)[0]

	g.NDChan[v1] = make(chan []byte, 1)
	g.NDChan[v2] = make(chan []byte, 1)

	g.n.RequestND(n1.Name, v1)
	g.n.RequestND(n2.Name, v2)

	ndByte1 := <-g.NDChan[v1]
	ndByte2 := <-g.NDChan[v2]

	var nd1, nd2 NeighbourDeegrees
	err := json.Unmarshal(ndByte1, &nd1)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(ndByte2, &nd2)
	if err != nil {
		panic(err)
	}

	minDist := 1000000000000
	for k, v := range nd1.Neighbours {
		if v2Dist, ok := nd2.Neighbours[k]; ok {
			minDist = int(math.Min(float64(minDist), float64(v+v2Dist)))
		}
	}
	for k, v := range nd2.Neighbours {
		if v1Dist, ok := nd1.Neighbours[k]; ok {
			minDist = int(math.Min(float64(minDist), float64(v+v1Dist)))
		}
	}

	delete(g.NDChan, v1)
	delete(g.NDChan, v2)

	return minDist
}