package main

import (
	bootstrap "version-checker/cmd"

	"go.uber.org/fx"
)

func main() {
	fx.New(bootstrap.Module).Run()
}
