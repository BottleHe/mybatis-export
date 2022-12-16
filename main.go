package main

import (
	"mybatis-export/cmd"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(1)
	cmd.Execute()
}
