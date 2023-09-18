package api

type Page struct {
	Num  int
	List []*PageUser
}

func NewPage() *Page {
	return &Page{0, make([]*PageUser, 0)}
}
func (p *Page) AddPageUser(list []*PageUser) {
	p.List = append(p.List, list...)
}
func (p *Page) AddPage(page *Page) {
	p.List = append(p.List, page.List...)
	p.Num = p.Num + page.Num
}
func (p *Page) AddNum(num int) {
	p.Num = p.Num + num
}

type PageUser struct {
	UserName       string
	MachineAddress string
	CreateTime     string
	MachineId      string
}

type GroupMsg struct {
	MachineAddress string
	Num            int32
	MachineId      string
}
