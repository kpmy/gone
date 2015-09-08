package model

import (
	"fmt"
)

type Link struct {
	NodeId int
	LinkId int
}

type Node struct {
	In      map[Link]interface{}
	Out     []interface{}
	Weights []int
}

func (n *Node) String() string {
	return fmt.Sprint("<", n.In, "->", len(n.Out), ">")
}

type Layer struct {
	Nodes []*Node
	Next  *Layer
}

func (l *Layer) String() string {
	return fmt.Sprint(l.Nodes, l.Next)
}

func NewNode() *Node {
	n := &Node{}
	n.In = make(map[Link]interface{})
	return n
}
