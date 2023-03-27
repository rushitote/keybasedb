package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PNode struct {
	port       string     // The port that this node is listening on
	memberList []string   // The list of nodes that this node knows about
	suspect    bool       // Whether this node is suspecting some other node to be dead
	mu         sync.Mutex // Mutex to lock the memberList
}

type PNodeHandler struct {
	node *PNode
}

func getExistingNodes(port string) []string {
	resp, err := http.Get("http://localhost:" + port + "/nodes")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var nodes []string
	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		nodes = strings.Split(string(b), ",")
	}
	return nodes
}

func NewPNode(port string, runningPort string) (p *PNode) {
	n := PNode{port: port, memberList: []string{}}
	if runningPort != "" {
		n.memberList = getExistingNodes(runningPort)
	}
	n.memberList = append(n.memberList, port)

	return &n
}

func (p *PNode) SendDiscoveryMessage() {
	// Wait for the server to start
	time.Sleep(3 * time.Second)
	p.mu.Lock()
	println("discovering")
	for _, node := range p.memberList {
		if node != p.port {
			var sendReq func(node string, tries int)
			sendReq = func(node string, tries int) {
				println("sending")
				if tries == 0 {
					return
				}
				client := &http.Client{}
				req, _ := http.NewRequest("GET", "http://localhost:"+node+"/add", nil)
				req.Header.Add("Node", p.port)
				_, err := client.Do(req)
				println("got response")
				if err != nil {
					sendReq(node, tries-1)
				}
			}
			sendReq(node, 3)
		}
	}
	println("discovery done")
	p.mu.Unlock()
	go p.CheckForRandomNodeFailure()
}

func (h *PNodeHandler) GetAllMemberNodes(w http.ResponseWriter, r *http.Request) {
	h.node.mu.Lock()
	nodes := strings.Join(h.node.memberList, ",")
	h.node.mu.Unlock()
	w.Write([]byte(nodes))
}

func (h *PNodeHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func verifyAlive(node string) bool {
	resp, err := http.Get("http://localhost:" + node + "/ping")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

func (h *PNodeHandler) AddMemberNode(w http.ResponseWriter, r *http.Request) {
	println("Request received to add", r.Header.Get("Node"))
	node := r.Header.Get("Node")
	if verifyAlive(node) {
		println("waiting for lock")
		h.node.mu.Lock()
		println("got lock")
		h.node.memberList = append(h.node.memberList, node)
		h.node.mu.Unlock()
	}
}

func (h *PNodeHandler) RemoveMemberNode(w http.ResponseWriter, r *http.Request) {
	println("Request received to remove", r.Header.Get("Node"))
	node := r.Header.Get("Node")
	go func() {
		h.node.mu.Lock()
		defer h.node.mu.Unlock()
		for i, n := range h.node.memberList {
			if n == node {
				h.node.memberList = append(h.node.memberList[:i], h.node.memberList[i+1:]...)
				return
			}
		}
	}()
}

func (h *PNodeHandler) CheckSuspectNode(w http.ResponseWriter, r *http.Request) {
	println("Request received to check", r.Header.Get("Node"))
	node := r.Header.Get("Node")

	var sendReq func(node string, tries int)
	sendReq = func(node string, tries int) {
		if tries == 0 {
			return
		}
		if verifyAlive(node) {
			client := &http.Client{}
			req, _ := http.NewRequest("POST", "http://localhost:"+r.Header.Get("From")+"/check-suspect-ack", nil)
			_, err := client.Do(req)
			if err != nil {
				sendReq(node, tries-1)
			}
		}
	}
	go sendReq(node, 1)
}

func (h *PNodeHandler) CheckSuspectNodeAck(w http.ResponseWriter, r *http.Request) {
	h.node.suspect = false
}

func (p *PNode) CheckForRandomNodeFailure() {
	// Wait T seconds every time
	time.Sleep(WaitBeforeCheckTime)
	p.mu.Lock()
	println("Currently has", len(p.memberList), "nodes")
	// Pick a random node
	if len(p.memberList) == 1 {
		p.mu.Unlock()
		p.CheckForRandomNodeFailure()
		return
	}

	rNum := rand.Intn(len(p.memberList) - 1)
	i := 0
	bef := false
	for i <= rNum {
		if p.port == p.memberList[i] {
			bef = true
		}
		i++
	}
	if bef {
		rNum++
	}

	checkedNode := p.memberList[rNum]

	println("Randomly checking node", checkedNode, "for failure")

	if verifyAlive(checkedNode) {
		p.mu.Unlock()
		p.CheckForRandomNodeFailure()
		return
	}
	i = 0
	j := 0
	// randomize the list
	lst := make([]string, len(p.memberList))
	copy(lst, p.memberList)
	for i < len(lst) {
		r := rand.Intn(len(lst))
		lst[i], lst[r] = lst[r], lst[i]
		i++
	}

	p.suspect = true

	i = 0

	for i < len(lst) && j < KNodes {
		if lst[i] == checkedNode || lst[i] == p.port {
			i++
			continue
		}
		// ask node to check if it knows about the failed node
		var sendReq func(node string, tries int)
		sendReq = func(node string, tries int) {
			if tries == 0 {
				return
			}
			client := &http.Client{}
			req, _ := http.NewRequest("POST", "http://localhost:"+node+"/check-suspect", nil)
			req.Header.Add("Node", checkedNode)
			req.Header.Add("From", p.port)
			_, err := client.Do(req)
			if err != nil {
				sendReq(node, tries-1)
			}
		}
		sendReq(lst[i], 1)
		j++
		i++
	}

	time.Sleep(WaitBeforeCheckTime)
	if p.suspect {
		// remove the node from the list
		// send a message to all nodes to remove the failed node
		for _, node := range lst {
			if node != p.port && node != checkedNode {
				var sendReq func(node string, tries int)
				sendReq = func(node string, tries int) {
					if tries == 0 {
						return
					}
					client := &http.Client{}
					req, _ := http.NewRequest("POST", "http://localhost:"+node+"/remove", nil)
					req.Header.Add("Node", checkedNode)
					_, err := client.Do(req)
					if err != nil {
						sendReq(node, tries-1)
					}
				}
				sendReq(node, 1)
			}
		}
		println("Removed", checkedNode, "from the list")
		p.memberList = append(p.memberList[:rNum], p.memberList[rNum+1:]...)
	}
	p.suspect = false
	p.mu.Unlock()
	p.CheckForRandomNodeFailure()
}

const (
	WaitBeforeCheckTime = 3 * time.Second
	KNodes              = 3
)
