package main

type User struct {
	ID   int64
	Name string
	Addr *Address
}

type Address struct {
	City   string
	ZIP    int
	LatLng [2]float64
}

var frank, lucy = User{}, User{}

func test() {
}
