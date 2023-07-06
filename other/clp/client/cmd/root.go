package cmd

import (
	"clp_client/metadata"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	MasterAddr      string
	MetadataAddr    string
	metadataUsr     string
	MetadataService metadata.MetadataService
)

var CmdRoot = &cobra.Command{
	Use: "clp_client",
}

func init() {
	CmdRoot.PersistentFlags().StringVar(&MasterAddr, "master_addr", "", "master address, in format of ip:port")
	CmdRoot.PersistentFlags().StringVar(&MetadataAddr, "metadata_addr", "", "metadata address, in format of ip:port")
	CmdRoot.PersistentFlags().StringVar(&metadataUsr, "metadata_usr", "", "metadata username")
	CmdRoot.AddCommand(CmdSearch)
	CmdRoot.AddCommand(CmdCompress)
}

func connectMetadataServer() {
	// Connect to metadata service
	MetadataService = &metadata.MetaMySQL{}
	for {
		var pwd string
		fmt.Print("Password: ")
		fmt.Scanln(&pwd)
		err := MetadataService.Connect(metadataUsr, pwd, MetadataAddr)
		if err == nil {
			break
		}
	}
}
