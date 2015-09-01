package main

import (
	"bytes"
	"encoding/json"
	"github.com/kpmy/ypk/halt"
	"gone"
	"io"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	log.SetFlags(0)
}

func pre(p *gone.Perc) {
	const (
		Nsens  = 25
		Nass   = 9
		Nreact = 2
	)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	p.Sens = nil
	var socache []*gone.Out
	for i := 0; i < Nsens; i++ {
		next := &gone.Sens{Id: "s." + strconv.Itoa(i)}
		outs := make([]*gone.Out, rnd.Intn(Nass)+1)
		for j := 0; j < len(outs); j++ {
			w := 0
			if rnd.Intn(2) == 0 {
				w = 1
			} else {
				w = -1
			}
			outs[j] = &gone.Out{Id: strconv.Itoa(j), Owner: next.Id, Weight: w}
			next.Out = outs
		}
		socache = append(socache, outs...)
		p.Sens = append(p.Sens, next)
	}
	p.Ass = nil
	var aocache []*gone.Out
	for i := 0; i < Nass; i++ {
		next := &gone.Ass{Id: "a." + strconv.Itoa(i)}
		next.In = &gone.In{}
		outs := make([]*gone.Out, rnd.Intn(Nreact)+1)
		for j := 0; j < len(outs); j++ {
			outs[j] = &gone.Out{Id: strconv.Itoa(j), Owner: next.Id, Weight: rnd.Intn(201) - 100}
			next.Out = outs
		}
		aocache = append(aocache, outs...)
		p.Ass = append(p.Ass, next)
	}
	{
		acache := make(map[string]map[string]interface{})
		j := 0
		for i := 0; i < len(socache); {
			r := socache[i]
			skip := false
			if _, ok := acache[p.Ass[j].Id]; ok {
				if _, ok := acache[p.Ass[j].Id][r.Owner]; ok {
					skip = true
					j++
					if j == len(p.Ass) {
						j = 0
					}
				}
			}
			if !skip {
				p.Ass[j].In.Io = append(p.Ass[j].In.Io, r)
				if _, ok := acache[p.Ass[j].Id]; !ok {
					acache[p.Ass[j].Id] = make(map[string]interface{})
				}
				acache[p.Ass[j].Id][r.Owner] = r
				i++
			}
		}
	}
	p.React = nil
	for i := 0; i < Nreact; i++ {
		next := &gone.React{Id: "r." + strconv.Itoa(i)}
		next.In = &gone.In{}
		next.Out = &gone.Out{Owner: next.Id, Id: "0"}
		p.React = append(p.React, next)
	}
	{
		rcache := make(map[string]map[string]interface{})
		j := 0
		for i := 0; i < len(aocache); {
			r := aocache[i]
			skip := false
			if _, ok := rcache[p.React[j].Id]; ok {
				if _, ok := rcache[p.React[j].Id][r.Owner]; ok {
					skip = true
					j++
					if j == len(p.React) {
						j = 0
					}
				}
			}
			if !skip {
				p.React[j].In.Io = append(p.React[j].In.Io, r)
				if _, ok := rcache[p.React[j].Id]; !ok {
					rcache[p.React[j].Id] = make(map[string]interface{})
				}
				rcache[p.React[j].Id][r.Owner] = r
				i++
			}
		}
	}

	var dump func(*gone.In, string)
	dump = func(in *gone.In, indent string) {
		for _, r := range in.Io {
			log.Println(indent, r.Owner+":"+r.Id)
		}
	}

	for _, r := range p.Ass {
		log.Println(r.Id)
		dump(r.In, "  ")
	}

	for _, r := range p.React {
		log.Println(r.Id)
		dump(r.In, "  ")
	}
}

func do(m map[string]*gone.Node, data []int) (ret map[string]int) {
	dataFor := func(id string) int {
		if x, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSuffix(id, ":0"), "s.")); err == nil {
			return data[x]
		} else {
			halt.As(100, err)
		}
		panic(0)
	}
	type pack struct {
		id string
		c  int
	}
	wg := new(sync.WaitGroup)
	rg := new(sync.WaitGroup)
	ret = make(map[string]int)
	res := make(chan pack)
	wg.Add(1)

	go func() {
		for c := range res {
			ret[c.id] = c.c
		}
		wg.Done()
	}()

	for _, n := range m {
		switch o := n.Obj.(type) {
		case *gone.Sens:
			wg.Add(1)
			go func(n *gone.Node, c int) {
				n.In[0].Write(c)
				wg.Done()
			}(n, dataFor(o.Id))
		case *gone.React:
			wg.Add(1)
			rg.Add(1)
			go func(n *gone.Node) {
				res <- pack{id: o.Id, c: n.Out[0].Read()}
				wg.Done()
				rg.Done()
			}(n)
		case *gone.Ass:
			//do nothing
		default:
			halt.As(100, reflect.TypeOf(o))
		}
	}
	go func() {
		rg.Wait()
		close(res)
	}()
	wg.Wait()
	return
}

func main() {
	p := &gone.Perc{}
	if f, err := os.Open("demo-0.pc"); err == nil {
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, f)
		json.Unmarshal(buf.Bytes(), p)
		f.Close()
	} else {
		pre(p)
	}
	//pre(p) //debug only

	var m interface{}
	m = do(gone.Load(p), []int{
		0, 0, 1, 0, 0,
		0, 0, 1, 0, 0,
		1, 1, 1, 1, 1,
		0, 0, 1, 0, 0,
		0, 0, 1, 0, 0})
	log.Println("+", m)

	m = do(gone.Load(p), []int{
		1, 0, 0, 0, 1,
		0, 1, 0, 1, 0,
		0, 0, 1, 0, 0,
		0, 1, 0, 1, 0,
		1, 0, 0, 0, 1})

	log.Println("x", m)
	if f, err := os.Create("demo-0.pc"); err == nil {
		buf, _ := json.MarshalIndent(p, "", " ")
		f.Write(buf)
		f.Close()
	}
}
