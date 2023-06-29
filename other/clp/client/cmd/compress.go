package cmd

import (
	"clp_client/weed"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	clpPath       string
	filerAddr     string
	compressLevel uint
)

var CmdCompress = &cobra.Command{
	Use:   "compress <tag> <dir> [--filer ipAddr] [--clp_path clpPath]",
	Short: "Compress logs to SeaweedFS.",
	Long:  "Compress the logs in a directory locally or on SeaweedFS.",
	Args:  cobra.ExactArgs(2),
	Run:   compress,
}

func init() {
	CmdSearch.Flags().StringVar(&clpPath, "clp_path", "", "path to clp binary")
	CmdSearch.Flags().StringVar(&filerAddr, "filer", "", "ip address of the filer")
	CmdSearch.Flags().UintVar(&compressLevel, "compress_level", 3, "Compression level 1-9. 1 runs fastest with low compression rate. 9 runs slowest with high compression rate.")
}

// Invoking clp to compress the logs.
func compressLog(targetPath string, resultPath string) error {
	cmd := exec.Command(clpPath, resultPath, targetPath, "--compression-level", fmt.Sprintf("%v", compressLevel))
	return cmd.Run()
}

func downloadDir(remotePath string, localPath string) error {
	return nil
}

func uploadArchive(archiveDir string) {
	// Get number of files
	dir, err := os.Open(archiveDir)
	if err != nil {
		log.Fatalln("Read archive dir fails.")
	}
	defer dir.Close()
	segmentDir, err := os.Open(filepath.Join(archiveDir, "s"))
	if err != nil {
		log.Fatalln("Read archive segment dir fails.")
	}
	defer segmentDir.Close()
	segments, err := segmentDir.Readdirnames(0)
	if err != nil {
		log.Fatalln("Walk archive segment dir fails.")
	}
	numFiles := 7 + len(segments)
	// Get a NeedleID from master
	key, err := weed.AssignFileKey(MasterAddr, numFiles)
	if err != nil {
		log.Fatalln("Get new NeedleID fails.")
	}
	// Upload files
	// TODO
}

func compress(cmd *cobra.Command, args []string) {
	// Parse Argument
	tag := args[0]
	dir := args[1]
	if tag == "" {
		log.Println("tag must not be empty.")
		return
	}
	if dir == "" {
		log.Println("log directory path must not be empty.")
		return
	}

	// Download file from filer if necessary
	if filerAddr != "" {
		localDir := fmt.Sprintf("/tmp/logweed/uncompressed/%v", time.Now())
		err := downloadDir(dir, localDir)
		if err != nil {
			return
		}
		dir = localDir
	}

	// Run compression
	compressedDir := fmt.Sprintf("/tmp/logweed/compressed/%v", time.Now())
	err := compressLog(dir, compressedDir)
	if err != nil {
		log.Println("Clp compress fails.")
		return
	}

}
