package gone

type Out struct {
	Owner  string
	Id     string
	Weight int
}

func (o *Out) String() string {
	return o.Owner + ":" + o.Id
}

type In struct {
	Io []*Out
}

type Sens struct {
	Id  string
	Out []*Out
}

type Ass struct {
	Id  string
	In  *In
	Out []*Out
}

type React struct {
	Id  string
	In  *In
	Out *Out
}

type Perc struct {
	Sens  []*Sens
	Ass   []*Ass
	React []*React
}
