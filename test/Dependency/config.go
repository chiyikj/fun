package Dependency

import "fmt"

type Config struct {
	Page int8
	T    T
}

type T struct {
	Name string
}

func (config *Config) New() {
	fmt.Println("我是config的new")
	config.Page = 5
}

type X struct {
	Name   string
	Config *Config `fun:"auto"`
}

func (x *X) New() {
	fmt.Println("我是x的new")
	x.Name = "x"
}
