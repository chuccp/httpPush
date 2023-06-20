package cluster

import (
	"log"
	"testing"
)

type AA interface {
	A()
}
type Person struct {
	name string
	age  int
}

func (Person *Person) A() {

}

type Student struct {
	*Person
	grade int
}

func TestType(t *testing.T) {

	s := &Student{}
	AAA(s)

}
func AAA(v any) {
	z, ok := v.(*Person)
	log.Println(z, ok)
}
