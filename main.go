package main

import (
	"os"
	"strconv"
)

func seed(args []string) {
	var nodes []*NodeInfo
	for i := 0; i < len(args); i++ {
		port, err := strconv.Atoi(args[i])
		if err != nil {
			panic(err)
		}
		nodes = append(nodes, &NodeInfo{
			Name:    "n" + args[i],
			Addr:    "0.0.0.0",
			Port:    args[i],
			APIPort: strconv.Itoa(1000 + port),
		})
	}
	cfg := CreateConfig(
		THREE,
		QUORUM,
		nodes)
	n := StartNode(cfg, nodes[0], nil)
	n.Server.Start()
}

func join(args []string) {
	var nodes []*NodeInfo
	for i := 0; i < len(args); i++ {
		port, err := strconv.Atoi(args[i])
		if err != nil {
			panic(err)
		}
		nodes = append(nodes, &NodeInfo{
			Name:    "n" + args[i],
			Addr:    "0.0.0.0",
			Port:    args[i],
			APIPort: strconv.Itoa(1000 + port),
		})
	}
	n := StartNode(nil, nodes[0], nodes[1])
	n.Server.Start()
}

func main() {
	args := os.Args
	args = args[1:]
	mode := args[0]
	switch mode {
	case "seed":
		seed(args[1:])
	case "join":
		join(args[1:])
	}
}
