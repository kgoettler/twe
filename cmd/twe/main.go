/*
Copyright © 2024 Ken Goettler <goettlek@gmail.com>
*/
package main

import "github.com/kgoettler/twe/cmd/twe/cmd"

var Version string = "dev"

func main() {
	cmd.RootCmd.Version = Version
	cmd.Execute()
}
