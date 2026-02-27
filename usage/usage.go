package main

import (
	"github.com/ihiiro/clog"
)

func main() {
	clog.Init() // sets up logfiles according to .env if exists

	clog.Printf("hello clog %d", 4)
	clog.Println("hello there partner clog")
	clog.Fatal("fatal clog")
}