package ares

import "time"

type Node struct {
	name       string
	updateTime int64

	cmd chan string
}

var HeartBeat_Interval = 20

func newNode(name string) *Node {
	node := &Node{
		name:       name,
		updateTime: 0,

		cmd: make(chan string),
	}
	node.onHeartBeat()
	return node
}

func (n *Node) onHeartBeat() {
	t := time.Now()
	updateTime := t.Unix()
	n.updateTime = updateTime
}

func (n *Node) isActive() bool {
	t := time.Now()
	updateTime := t.Unix()

	interval := updateTime - n.updateTime
	if interval > (int64)(HeartBeat_Interval) {
		return false
	}
	return true
}

func (n *Node) sendCmd(appName string) {
	n.cmd <- appName
}
