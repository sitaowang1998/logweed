package command

import "github.com/spf13/cobra"

var (
	bts_str string
	ets_str string
)

var CmdSearch = &cobra.Command{
	Use:   "search <tag> <query> [--bts bts] [--ets ets]",
	Short: "Search the logs in SeaweedFS.",
	Long:  "Search the log files associated with the tag.",
	Args:  cobra.ExactArgs(2),
	Run:   search,
}

func init() {
	CmdSearch.Flags().StringVar(&bts_str, "bts", "", "begin timestamp")
	CmdSearch.Flags().StringVar(&ets_str, "ets", "", "end timestamp")
}

func search(cmd *cobra.Command, args []string) {

}
