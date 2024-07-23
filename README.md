
fillstruct - fills a struct literal with default values

---

For example, given the following types,

```golang
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

```golang
var frank = User{}
```

becomes:

```golang
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

```sh
% go install github.com/CaiJinKen/fillstruct@v0.2.2
```

## Usage

```sh
% fillstruct -file=<filename> -line=<line number> -writeback=true
or
% fillstruct -file <filename> -line <line number> -writeback
```

Flags:

```sh
-file string
    filename
-line int
    line number of the struct literal
-only-changed
    just print changed line, false will print all info
-std-out
    print info into stdout (default true)
-version string
    print fillstruct version
-writeback
    writeback to the file
```

If -offset as well as -line are present, then the tool first uses the
more specific offset information. If there was no struct literal found
at the given offset, then the line information is used.

what types of assign statement supported? You can find use case in [test.go](https://github.com/CaiJinKen/fillstruct/blob/master/test.go) for detail.

- [x] global variable
- [x] general local variable
- [x] local variable in local function
- [x] local variable in function of a return statement

### Vim / Neovim ?

sure! You can add [vim-fillstruct](https://github.com/CaiJinKen/vim-fillstruct) plugin in vim/neovim
