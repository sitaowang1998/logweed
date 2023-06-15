package main

import (
	"clp_client/command"
	"github.com/spf13/cobra"
)

var (
	master_addr   string
	metadata_addr string
)

func main() {
	var cmdRoot = &cobra.Command{Use: "clp_client"}

	cmdRoot.PersistentFlags().StringVar(&master_addr, "master_addr", "", "master address, in format of ip:port")
	cmdRoot.AddCommand(command.CmdSearch)

	err := cmdRoot.Execute()
	if err != nil {
		println("Fail to run clp_client. %v", err)
		return
	}
}
