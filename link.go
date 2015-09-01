package gone

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/halt"
	"log"
	"math"
	"reflect"
)

type Link struct {
	i      chan int
	o      chan int
	owner  string
	id     string
	Weight int
}

type Node struct {
	Out []*Link
	In  []*Link
	Obj interface{}
}

var debug bool

func init() {
	debug = false
}

func (l *Link) Write(c int) {
	if debug {
		log.Println("write", l.owner+":"+l.id, "<-", c)
	}
	l.i <- c
	if debug {
		log.Println("wrote", l.owner+":"+l.id, "<-", c)
	}
}

func (l *Link) Read() int {
	if debug {
		defer log.Println("ready", l.owner+":"+l.id)
		log.Println("read", l.owner+":"+l.id)
	}
	return <-l.o
}

func NewLink(owner, id string, weight int) (ret *Link) {
	ret = &Link{owner: owner, id: id, Weight: weight}
	ret.i = make(chan int)
	ret.o = make(chan int)
	go func(t *Link) {
		x := <-t.i
		for {
			t.o <- x
		}
	}(ret)
	return
}

func Load(p *Perc) (ret map[string]*Node) {
	this := func(owner string) (ret interface{}) {
		switch owner[0] {
		case 'a':
			for _, x := range p.Ass {
				if x.Id == owner {
					return x
				}
			}
		case 's':
			for _, x := range p.Sens {
				if x.Id == owner {
					return x
				}
			}
		default:
			halt.As(100, owner)
		}
		return
	}
	var run func(_x interface{}, id string) *Link
	run = func(_x interface{}, id string) *Link {
		switch x := _x.(type) {
		case *Ass:
			cache := make(map[string]*Link)
			if n, ok := ret[x.Id]; !ok {
				next := new(Node)
				for _, k := range x.In.Io {
					next.In = append(next.In, run(this(k.Owner), k.Id))
				}
				for _, o := range x.Out {
					l := NewLink(x.Id, o.Id, o.Weight)
					cache[o.Id] = l
					next.Out = append(next.Out, l)
				}
				ret[x.Id] = next
				next.Obj = x
				go func(n *Node) {
					for {
						var buf []int
						for _, k := range n.In {
							buf = append(buf, k.Read())
						}
						var sum int
						for _, c := range buf {
							sum += c
						}
						for _, k := range n.Out {
							if sum > 1 {
								k.Write(k.Weight)
							} else {
								k.Write(0)
							}
						}
					}
				}(next)
			} else {
				for _, o := range n.Out {
					cache[o.id] = o
				}
			}
			assert.For(len(cache) > 0, 60)
			return cache[id]
		case *Sens:
			cache := make(map[string]*Link)
			if n, ok := ret[x.Id]; !ok {
				next := new(Node)
				for _, o := range x.Out {
					l := NewLink(x.Id, o.Id, o.Weight)
					cache[o.Id] = l
					next.Out = append(next.Out, l)
				}
				next.In = append(next.In, NewLink(x.Id, "0", 0))
				ret[x.Id] = next
				next.Obj = x
				go func(n *Node) {
					x := n.In[0].Read()
					for _, k := range n.Out {
						if x == 1 {
							k.Write(k.Weight)
						} else {
							k.Write(0)
						}
					}
				}(next)
			} else {
				for _, o := range n.Out {
					cache[o.id] = o
				}
			}
			assert.For(len(cache) > 0, 60)
			return cache[id]
		default:
			halt.As(100, reflect.TypeOf(x))
		}
		panic(0)
	}
	ret = make(map[string]*Node)
	for _, r := range p.React {
		next := new(Node)
		next.Out = append(next.Out, NewLink(r.Id, "0", 0))
		for _, k := range r.In.Io {
			next.In = append(next.In, run(this(k.Owner), k.Id))
		}
		next.Obj = r
		ret[r.Id] = next
		go func(n *Node) {
			for {
				var buf []int
				for _, k := range n.In {
					buf = append(buf, k.Read())
				}
				var sum int
				for _, c := range buf {
					sum += c
				}
				for _, k := range n.Out {
					if math.Signbit(float64(sum)) {
						k.Write(1)
					} else {
						k.Write(-1)
					}
				}
			}
		}(next)
	}
	return
}
