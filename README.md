
fillstruct - fills a struct literal with default values

---

For example, given the following types,
```
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
```
the following struct literal
```
var frank = User{}
```
becomes:
```
var frank = User{
	ID:   0,
	Name: "",
	Addr: &Address{
		City: "",
		ZIP:  0,
		LatLng: [2]float64{
			0.0,
			0.0,
		},
	},
}
```
after applying fillstruct.

## Installation

```
% go install github.com/CaiJinKen/fillstruct
```

## Usage

```
% fillstruct [-modified] -file=<filename> -offset=<byte offset> -line=<line number>
```

Flags:

	-file:     filename
	-modified: read an archive of modified files from stdin
	-offset:   byte offset of the struct literal, optional if -line is present
	-line:     line number of the struct literal, optional if -offset is present

If -offset as well as -line are present, then the tool first uses the
more specific offset information. If there was no struct literal found
at the given offset, then the line information is used.
