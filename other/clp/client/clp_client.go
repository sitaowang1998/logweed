package main

import (
	"clp_client/cmd"
)

func main() {
	err := cmd.CmdRoot.Execute()
	if err != nil {
		println("Fail to run clp_client. %v", err)
		return
	}
}
