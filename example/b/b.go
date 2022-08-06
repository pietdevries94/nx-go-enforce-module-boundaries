package b

import (
	"fmt"

	"github.com/pietdevries94/nx-go-enforce-module-boundaries/example/a"
)

func PrintHello() {
	fmt.Println(a.GetHello())
}
