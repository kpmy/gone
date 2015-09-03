package main

import (
	"github.com/kpmy/ypk/assert"
	"gone/perc"
	"log"
	"math/rand"
)

const (
	Nsens      = 256 * 256
	MaxSensOut = 128
	Nass       = 256
	Nreact     = 16
)

func init() {
	assert.For(MaxSensOut <= Nass, 20)
	log.SetFlags(0)
}

func New(nn, no int, w_fn func() int) (ret *model.Layer) {
	ret = &model.Layer{}
	for i := 0; i < nn; i++ {
		next := model.NewNode()
		next.Out = make([]interface{}, rand.Intn(no)+1)
		next.Weights = make([]int, len(next.Out))
		if w_fn != nil {
			for i := 0; i < len(next.Weights); i++ {
				next.Weights[i] = w_fn()
			}
		}
		ret.Nodes = append(ret.Nodes, next)
	}
	return
}

func Join(in *model.Layer, out *model.Layer) {
	j := 0
	type im map[int]interface{}
	cache := make([]im, len(out.Nodes))
	for k, n := range in.Nodes {
		for i := 0; i < len(n.Out); {
			assert.For(len(n.Out) <= len(out.Nodes), 20, len(n.Out), len(out.Nodes))
			l := model.Link{NodeId: k, LinkId: i}
			skip := false
			if cache[j] != nil {
				if _, ok := cache[j][k]; ok {
					skip = true
					j++
					if j == len(cache) {
						j = 0
					}
				}
			}
			if !skip {
				out.Nodes[j].In[l] = nil
				if cache[j] == nil {
					cache[j] = make(im)
				}
				cache[j][k] = l
				//log.Println(l, "to", j)
				i++
			}
			//log.Println(k, len(n.Out), i, j)
		}
	}
	in.Next = out
}

func main() {
	s := New(Nsens, MaxSensOut, func() int {
		x := rand.Intn(2)
		if x == 0 {
			return 1
		} else {
			return -1
		}
	})
	a := New(Nass, Nreact, func() int { return rand.Intn(201) - 100 })
	r := New(Nreact, 1, nil)

	Join(s, a)
	Join(a, r)
}
