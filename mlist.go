package main

import (
	"strconv"

	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
)

// MemberList is a wrapper around the memberlist package
type MemberList struct {
	List *memberlist.Memberlist
}

func CreateMemberList(node *NodeInfo, seedNode *NodeInfo, processMsg func(b []byte)) (m *MemberList) {
	port, err := strconv.Atoi(node.Port)
	if err != nil {
		panic(err)
	}

	config := memberlist.DefaultLocalConfig()
	config.BindPort = port
	config.AdvertisePort = port
	config.Name = node.Name
	config.Delegate = &MemberListDelegate{
		ProcessMsg: processMsg,
	}
	config.LogOutput = logrus.StandardLogger().WriterLevel(logrus.DebugLevel)

	list, err := memberlist.Create(config)
	if err != nil {
		panic(err)
	}

	if seedNode != nil {
		addr := seedNode.Addr + ":" + seedNode.Port
		_, err := list.Join([]string{addr})
		if err != nil {
			panic(err)
		}
	}

	m = &MemberList{
		List: list,
	}

	return m
}

func (m *MemberList) CheckIfNodeAlive(node *NodeInfo) bool {
	for _, member := range m.List.Members() {
		if member.Name == node.Name {
			return true
		}
	}
	return false
}

func (m *MemberList) FindNode(name string) (node *memberlist.Node) {
	for _, member := range m.List.Members() {
		if member.Name == name || PadName(member.Name) == name {
			return member
		}
	}
	return nil
}

func (m *MemberList) SendTCP(msg []byte, name string) {
	node := m.FindNode(name)
	if node != nil {
		m.List.SendReliable(node, msg)
	}
}

func (m *MemberList) SendUDP(msg []byte, name string) {
	node := m.FindNode(name)
	if node != nil {
		m.List.SendBestEffort(node, msg)
	}
}

type MemberListDelegate struct {
	ProcessMsg func([]byte)
}

// TODO: Implement these methods

func (d *MemberListDelegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *MemberListDelegate) NotifyMsg(b []byte) {
	d.ProcessMsg(b)
}

func (d *MemberListDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

func (d *MemberListDelegate) LocalState(join bool) []byte {
	return []byte{}
}

func (d *MemberListDelegate) MergeRemoteState(buf []byte, join bool) {
	println("MergeRemoteState", string(buf))
}

func PadName(name string) string {
	for len(name) < 8 {
		name = "0" + name
	}
	return name
}
