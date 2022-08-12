package b

import (
	"fmt"

	"github.com/pietdevries94/nx-go-enforce-module-boundaries/example/a"
	"github.com/pietdevries94/nx-go-enforce-module-boundaries/example/c"
)

func PrintHello() {
	fmt.Println(a.GetHello(), c.GetHello())
}
