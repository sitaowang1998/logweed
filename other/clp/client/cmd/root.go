package cmd

import (
	"clp_client/metadata"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"syscall"
)

var (
	MasterAddr      string
	MetadataAddr    string
	metadataUsr     string
	metadataPwd     string
	MetadataService metadata.MetadataService
)

var CmdRoot = &cobra.Command{
	Use: "clp_client",
}

func init() {
	CmdRoot.PersistentFlags().StringVar(&MasterAddr, "master_addr", "", "master address, in format of ip:port")
	CmdRoot.PersistentFlags().StringVar(&MetadataAddr, "metadata_addr", "", "metadata address, in format of ip:port")
	CmdRoot.PersistentFlags().StringVar(&metadataUsr, "metadata_usr", "", "metadata username")
	CmdRoot.PersistentFlags().StringVar(&metadataPwd, "metadata_pwd", "", "metadata password")
	CmdRoot.AddCommand(CmdSearch)
	CmdRoot.AddCommand(CmdCompress)
}

func connectMetadataServer() {
	// Connect to metadata service
	MetadataService = &metadata.MetaMySQL{}
	if len(metadataPwd) == 0 {
		for {
			fmt.Print("Password: ")
			pwd, err := terminal.ReadPassword(syscall.Stdin)
			if err != nil {
				continue
			}
			err = MetadataService.Connect(metadataUsr, string(pwd), MetadataAddr)
			if err == nil {
				break
			}
		}
	} else {
		err := MetadataService.Connect(metadataUsr, metadataPwd, MetadataAddr)
		if err != nil {
			log.Println("Wrong metadata password.")
			syscall.Exit(1)
		}
	}
}
