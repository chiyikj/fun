package main

import "fun"
import _ "fun/test/service/userService"

func main() {
	fun.Gen()
	fun.Start(3000)
}
