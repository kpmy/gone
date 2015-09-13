package proc

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/fn"
	"github.com/kpmy/ypk/halt"
	"gone/model"
	"sync"
)

const (
	Sensoric    = "sensor"
	Associative = "assoc"
	Reactive    = "react"
)

var fm map[string]func(*model.Node)

var defLink model.Link = model.Link{NodeId: 0, LinkId: 0}

func init() {
	fm = make(map[string]func(*model.Node))

	fm[Sensoric] = func(n *model.Node) {
		v := n.In[defLink].(int)
		for i, w := range n.Weights {
			n.Out[i] = Sfunc(w, v)
		}
	}

	fm[Associative] = func(n *model.Node) {
		var vl []int
		for _, v := range n.In {
			vl = append(vl, v.(int))
		}
		v := Afunc(vl...)
		for i, w := range n.Weights {
			n.Out[i] = Afuncw(w, v)
		}
	}
	fm[Reactive] = func(n *model.Node) {
		var vl []int
		for _, v := range n.In {
			vl = append(vl, v.(int))
		}
		n.Out[0] = Rfunc(vl...)
	}
}

func Sfunc(weight, signal int) int {
	return weight * signal
}

func Afunc(signal ...int) int {
	var r int
	for _, x := range signal {
		r = r + x
	}
	if r >= 8 {
		r = 1
	} else {
		r = 0
	}
	return r
}

func Afuncw(weight, signal int) int {
	return weight * signal
}

func Rfunc(signal ...int) int {
	//log.Println(signal)
	var r int
	for _, x := range signal {
		r = r + x
	}
	if r > 0 {
		r = 1
	} else if r < 0 {
		r = -1
	} else {
		r = 0
	}
	return r
}

func Process(l *model.Layer) {
	assert.For(l != nil, 20)
	if fn, ok := fm[l.Role]; ok {
		wg := new(sync.WaitGroup)
		for _, n := range l.Nodes {
			wg.Add(1)
			go func(n *model.Node) {
				fn(n)
				wg.Done()
			}(n)
		}
		wg.Wait()
	} else {
		halt.As(100, l.Role)
	}
}

func Transfer(l *model.Layer) {
	assert.For(l != nil && l.Next != nil, 20)
	wg := new(sync.WaitGroup)
	for _, n := range l.Next.Nodes {
		wg.Add(1)
		go func(n *model.Node) {
			tmp := make(map[model.Link]interface{})
			for k, _ := range n.In {
				v := l.Nodes[k.NodeId].Out[k.LinkId]
				if fn.IsNil(v) {
					halt.As(100, k)
				}
				tmp[k] = v
			}
			n.In = tmp
			wg.Done()
		}(n)
	}
	wg.Wait()
}
