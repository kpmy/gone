package main

import (
	"github.com/kpmy/ypk/assert"
	"github.com/kpmy/ypk/fn"
	"gone/model"
	"gone/model/proc"
	"gone/model/std"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"time"
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

func generate() *model.Layer {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := std.New(proc.Sensoric, Nsens, MaxSensOut, func() int {
		x := rnd.Intn(2)
		if x == 0 {
			return 1
		} else {
			return -1
		}
	})
	a := std.New(proc.Associative, Nass, Nreact, func() int { return rnd.Intn(201) - 100 })
	r := std.New(proc.Reactive, Nreact, 1, nil)

	std.Join(s, a)
	std.Join(a, r)
	return s
}

const name = "pc"

func main() {
	var s, r *model.Layer
	if _, err := os.Stat(name); os.IsNotExist(err) {
		s = generate()

		if f, err := os.Create(name); err == nil {
			if err := std.Store(f, s); err != nil {
				log.Fatal(err)
			}
			f.Close()
		}
	} else {
		log.Println("file exists")
	}

	if f, err := os.Open(name); err == nil {
		defer f.Close()
		if s, err = std.Load(f); err != nil {
			log.Fatal(err)
		}
	}
	white := color.RGBA{R: 0xff, G: 0xFF, B: 0xFF, A: 0xFF}
	if fi, err := os.Open("0.png"); err == nil {
		if i, _, err := image.Decode(fi); err == nil {
			j := 0
			log.Println("image", i.Bounds())
			for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
				for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
					c := i.At(x, y)
					if c == white {
						s.Nodes[j].In[model.Link{NodeId: 0, LinkId: 0}] = 1
					} else {
						s.Nodes[j].In[model.Link{NodeId: 0, LinkId: 0}] = 0
					}
					j++
				}
			}

		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	/* for _, n := range s.Nodes {
		n.In[model.Link{NodeId: 0, LinkId: 0}] = rand.Intn(2)
	} */
	t := make([]int, Nreact)
	for i := 0; i < len(t); i++ {
		t[i] = -1
	}
	t[0] = 1

	for stop := false; !stop; { //autolearn for t pattern
		for l := s; !fn.IsNil(l); l = l.Next {
			proc.Process(l)
			if l.Next != nil {
				proc.Transfer(l)
			} else {
				r = l
			}
		}
		a := make([]int, Nreact)
		for i, n := range r.Nodes {
			a[i] = n.Out[0].(int)
		}
		log.Println(t)
		log.Println(a)
		stop = true
		for i, _ := range t {
			switch {
			case t[i] == 1 && a[i] == 1: //ok
			case t[i] == 1 && a[i] != t[i]:
				std.Update(s.Next, r.Nodes[i].In, 1)
				stop = false
			case t[i] == -1 && a[i] != t[i]:
				std.Update(s.Next, r.Nodes[i].In, -1)
				stop = false
			}
		}
	}

	if f, err := os.Create(name + ".0"); err == nil {
		if err := std.Store(f, s); err != nil {
			log.Fatal(err)
		}
		f.Close()
	}
}
