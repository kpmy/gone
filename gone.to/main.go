package main

import (
	"encoding/json"
	"github.com/kpmy/ypk/assert"
	"gone/perc"
	"gone/perc/std"
	"io"
	"log"
	"math/rand"
	"os"
)

const (
	Nsens      = 64 * 64
	MaxSensOut = 128
	Nass       = 256
	Nreact     = 16
)

func init() {
	assert.For(MaxSensOut <= Nass, 20)
	log.SetFlags(0)
}

type Link struct {
	NodeId int `json:"node"`
	LinkId int `json:"out"`
}

type NodeStruct struct {
	Outs []int  `json:"out"`
	Ins  []Link `json:"in,omitempty"`
}

type LayerStruct struct {
	Nodes []*NodeStruct `json:"node"`
}

func Store(wr io.Writer, first *model.Layer) error {
	res := make([]*LayerStruct, 0)
	for l := first; l != nil; l = l.Next {
		next := &LayerStruct{}
		for _, n := range l.Nodes {
			node := &NodeStruct{}
			for _, w := range n.Weights {
				node.Outs = append(node.Outs, w)
			}
			for k, _ := range n.In {
				link := Link{NodeId: k.NodeId, LinkId: k.LinkId}
				node.Ins = append(node.Ins, link)
			}
			next.Nodes = append(next.Nodes, node)
		}
		res = append(res, next)
	}
	return json.NewEncoder(wr).Encode(res)
}

func Load(rd io.Reader) (ret *model.Layer, err error) {
	ll := make([]LayerStruct, 0)
	if err = json.NewDecoder(rd).Decode(ll); err == nil {
		var this *model.Layer
		for _, l := range ll {
			if ret == nil {
				this = &model.Layer{}
				ret = this
			} else {
				this.Next = &model.Layer{}
				this = this.Next
			}
			for _, n := range l.Nodes {
				node := model.NewNode()
				for _, w := range n.Outs {
					node.Weights = append(node.Weights, w)
				}
				node.Out = make([]interface{}, len(node.Weights))
				for _, l := range n.Ins {
					link := model.Link{LinkId: l.LinkId, NodeId: l.NodeId}
					node.In[link] = nil
				}
				this.Nodes = append(this.Nodes, node)
			}
		}
	}
	return
}

func main() {
	s := std.New(Nsens, MaxSensOut, func() int {
		x := rand.Intn(2)
		if x == 0 {
			return 1
		} else {
			return -1
		}
	})
	a := std.New(Nass, Nreact, func() int { return rand.Intn(201) - 100 })
	r := std.New(Nreact, 1, nil)

	std.Join(s, a)
	std.Join(a, r)

	log.Println(s)

	if f, err := os.Create("0.pc"); err == nil {
		Store(f, s)
		f.Close()
	}

	if f, err := os.Open("0.pc"); err == nil {
		if s, err = Load(f); err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	log.Println(s)
}
