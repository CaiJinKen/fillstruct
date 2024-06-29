package main

import (
	"fmt"
	"time"

	"golang.org/x/tools/go/packages"
)

// User something
type User struct {
	ID   int64    `json:"id"`   // common 1
	Name string   `json:"name"` // common 2
	Addr *Address `json:"addr"` // common 3
}

type Address struct {
	City   string
	ZIP    int
	LatLng [2]float64
	List   *A
}
type A struct {
	name string
	addr []string
	ttt  time.Time
}

var frank = User{} // var frank = User{ID: 0, Name: "", Addr: &Address{City: "", ZIP: 0, LatLng: [2]float64{0.0, 0.0}, List: &A{name: "", addr: []string{}, ttt: time.Time{}}}}

var pkg = packages.Config{} // var pkg = packages.Config{Mode: 0, Context: nil, Logf: func(string, []interface{}) { panic("not implemented") }, Dir: "", Env: []string{}, BuildFlags: []string{}, Fset: &token.FileSet{}, ParseFile: func(*token.FileSet, string, []byte) (*ast.File, error) { panic("not implemented") }, Tests: false, Overlay: map[string][]byte{"": {}}}

func generalVariable() {
	bob, alice := User{}, User{}

	// bob, alice := User{ID: 0, Name: "", Addr: &Address{City: "", ZIP: 0, LatLng: [2]float64{0.0, 0.0}, List: &A{name: "", addr: []string{}, ttt: time.Time{}}}}, User{ID: 0, Name: "", Addr: &Address{City: "", ZIP: 0, LatLng: [2]float64{0.0, 0.0}, List: &A{name: "", addr: []string{}, ttt: time.Time{}}}}

	// gpk := packages.Config{Mode: 0, Context: nil, Logf: func(string, []interface{}) { panic("not implemented") }, Dir: "", Env: []string{}, BuildFlags: []string{}, Fset: &token.FileSet{}, ParseFile: func(*token.FileSet, string, []byte) (*ast.File, error) { panic("not implemented") }, Tests: false, Overlay: map[string][]byte{"": {}}}

	fmt.Println(bob, alice)
}

func localFunc() func() User {
	fn := func() User {
		u := User{} // u = User{ID: 0, Name: "", Addr: &Address{City: "", ZIP: 0, LatLng: [2]float64{0.0, 0.0}, List: &A{name: "", addr: []string{}, ttt: time.Time{}}}}
		return u
	}

	return fn
}

func returnFunc() func() User {
	return func() User {
		u := User{} // u := User{ID: 0, Name: "", Addr: &Address{City: "", ZIP: 0, LatLng: [2]float64{0.0, 0.0}, List: &A{name: "", addr: []string{}, ttt: time.Time{}}}}
		return u
	}
}
