package main

import (
	"net/http"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		panic("No port given")
	}

	port := args[0]
	runningPort := ""
	if len(args) > 1 {
		runningPort = args[1]
	}
	n := NewPNode(port, runningPort)
	h := PNodeHandler{node: n}
	http.HandleFunc("/nodes", h.GetAllMemberNodes)
	http.HandleFunc("/add", h.AddMemberNode)
	http.HandleFunc("/remove", h.RemoveMemberNode)
	http.HandleFunc("/check-suspect", h.CheckSuspectNode)
	http.HandleFunc("/check-suspect-ack", h.CheckSuspectNodeAck)
	println("Listening on port " + port)
	go h.node.SendDiscoveryMessage()
	http.ListenAndServe(":"+port, nil)
}
