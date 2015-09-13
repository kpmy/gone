package std

import (
	"encoding/json"
	"gone/model"
	"io"
)

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
	Role  string        `json:"role"`
}

func Store(wr io.Writer, first *model.Layer) error {
	res := make([]*LayerStruct, 0)
	for l := first; l != nil; l = l.Next {
		next := &LayerStruct{Role: l.Role}
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
	ll := make([]*LayerStruct, 0)
	if err = json.NewDecoder(rd).Decode(&ll); err == nil {
		var this *model.Layer
		for _, l := range ll {
			if ret == nil {
				this = &model.Layer{Role: l.Role}
				ret = this
			} else {
				this.Next = &model.Layer{Role: l.Role}
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
