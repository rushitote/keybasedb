package main

import (
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

func dump1() {
	args := os.Args[1:]
	port, err := strconv.Atoi(args[0])
	if err != nil {
		panic(err)
	}

	config := memberlist.DefaultLocalConfig()
	config.BindPort = port
	config.AdvertisePort = port
	config.Name = "node" + args[0]
	config.Delegate = &MemberListDelegate{
		ProcessMsg: func(b []byte) {
			println("ProcessMsg", string(b))
		},
	}

	list, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}

	if len(args) > 1 {
		addr := "0.0.0.0:" + args[1]
		_, err := list.Join([]string{addr})
		if err != nil {
			panic(err)
		}
	}

	for {
		for _, member := range list.Members() {
			println(member.Name, member.Addr.String(), member.Port)
		}
		println("----")
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	otherNode := NodeInfo{
		Name:    "n8001",
		Addr:    "0.0.0.0",
		Port:    "8001",
		APIPort: "9001",
	}
	ni := NodeInfo{
		Name:    "n8000",
		Addr:    "0.0.0.0",
		Port:    "8000",
		APIPort: "9000",
	}

	if len(os.Args) > 1 {
		c := 1
		n := StartNode(nil, &otherNode, &ni)
		for {
			println("running", n.Info.Name)
			time.Sleep(5 * time.Second)
			// if c%2 != 0 {
			// 	n.Write("key1", "value1")
			// } else {
			// 	n.Delete("key1")
			// }
			c++
		}
	}

	cfg := CreateConfig(
		TWO,
		ALL,
		[]*NodeInfo{&ni, &otherNode})

	n := StartNode(cfg, &ni, nil)

	n.Server.Start()

	// for {
	// 	v, err := n.Read("key1")
	// 	if err != nil {
	// 		println("err", err.Error())
	// 	}
	// 	println("value = ", v)
	// 	time.Sleep(3 * time.Second)
	// }
}
