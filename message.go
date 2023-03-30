package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

func (n *Node) ProcessMsg(b []byte) {
	mType := uint8(b[0])
	sender := string(b[1:9])
	msg := b[9:]
	if mType == REQUEST_CONFIG {
		n.processRequestConfig(sender)
	} else if mType == RESPONSE_CONFIG {
		n.processResponseConfig(msg)
	} else if mType == REQUEST_READ {
		n.processRequestRead(sender, msg)
	} else if mType == RESPONSE_READ {
		n.processResponseRead(msg)
	} else if mType == REQUEST_WRITE {
		n.processRequestWrite(sender, msg)
	} else if mType == RESPONSE_WRITE {
		n.processResponseWrite(msg)
	} else {
		log.Infof("Unknown message type %s", mType)
	}
}

func (n *Node) RequestConfig(seedNode *NodeInfo) {
	var b []byte
	b = append(b, REQUEST_CONFIG)
	b = append(b, []byte(n.Info.GetSenderName())...)
	log.Infof("Requesting config from %s", seedNode.Name)
	n.MList.SendTCP(b, seedNode.Name)
}

func (n *Node) processRequestConfig(sender string) {
	cfg := n.Config.SerializeConfig()
	var b []byte
	b = append(b, RESPONSE_CONFIG)
	b = append(b, []byte(n.Info.GetSenderName())...)
	b = append(b, cfg...)
	log.Infof("Sending config to %s", sender)
	n.MList.SendTCP(b, sender)
}

func (n *Node) processResponseConfig(msg []byte) {
	n.Config = DeserializeConfig(msg)
	n.Router = CreateRouter(n.Config)
}

func (n *Node) RequestRead(key string, to string) {
	var b []byte
	b = append(b, REQUEST_READ)
	b = append(b, []byte(n.Info.GetSenderName())...)
	b = append(b, []byte(key)...)
	log.Infof("Requesting read of key=%s from %s", key, to)
	n.MList.SendTCP(b, to)
}

func (n *Node) processRequestRead(sender string, msg []byte) {
	key := string(msg)
	value, err := n.Engine.Read(key)
	if err != nil {
		panic(err)
	}
	var b []byte
	b = append(b, RESPONSE_READ)
	b = append(b, []byte(n.Info.GetSenderName())...)
	respMsg, err := json.Marshal(ReadRequestMsg{key, value})
	if err != nil {
		panic(err)
	}
	b = append(b, respMsg...)
	log.Infof("Sending read response of key=%s to %s", key, sender)
	n.MList.SendTCP(b, sender)
}

func (n *Node) processResponseRead(msg []byte) {
	var respMsg ReadRequestMsg
	err := json.Unmarshal(msg, &respMsg)
	if err != nil {
		panic(err)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.opsChan[respMsg.Key] <- []byte(respMsg.Value)
}

func (n *Node) RequestWrite(key string, value string, to string) {
	var b []byte
	b = append(b, REQUEST_WRITE)
	b = append(b, []byte(n.Info.GetSenderName())...)
	reqMsg, err := json.Marshal(WriteRequestMsg{key, value})
	if err != nil {
		panic(err)
	}
	b = append(b, reqMsg...)
	log.Infof("Requesting write of key=%s to %s", key, to)
	n.MList.SendTCP(b, to)
}

func (n *Node) processRequestWrite(sender string, msg []byte) {
	var reqMsg WriteRequestMsg
	err := json.Unmarshal(msg, &reqMsg)
	if err != nil {
		panic(err)
	}

	prevVal, err := n.Engine.Read(reqMsg.Key)
	if err != nil {
		panic(err)
	}
	if prevVal == "" || GetTimestampFromValue(prevVal) < GetTimestampFromValue(reqMsg.Value) {
		n.Engine.Write(reqMsg.Key, reqMsg.Value)
	}
	var b []byte
	b = append(b, RESPONSE_WRITE)
	b = append(b, []byte(n.Info.GetSenderName())...)
	b = append(b, msg...)
	log.Infof("Sending write response of key=%s to %s", reqMsg.Key, sender)
	n.MList.SendTCP(b, sender)
}

func (n *Node) processResponseWrite(msg []byte) {
	var reqMsg WriteRequestMsg
	err := json.Unmarshal(msg, &reqMsg)
	if err != nil {
		panic(err)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.opsChan[reqMsg.Key] <- []byte(reqMsg.Value)
}

// Types of messages
const (
	REQUEST_CONFIG = iota
	RESPONSE_CONFIG
	REQUEST_READ
	REQUEST_WRITE
	RESPONSE_READ
	RESPONSE_WRITE
)

// TODO: find a better way to serialize/deserialize than json

type ReadRequestMsg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type WriteRequestMsg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
