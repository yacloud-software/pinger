package matrix

type networklist struct {
	networks []*networkdef
}
type networkdef struct {
	network string
	hosts   []string
}

func (nl *networklist) AddIP(ip string) {
}
