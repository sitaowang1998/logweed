package cmd

import (
	"clp_client/metadata"
	"clp_client/weed"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
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
	CmdCompress.Flags().StringVar(&clpPath, "clp_path", "", "path to clp binary")
	CmdCompress.Flags().StringVar(&filerAddr, "filer", "", "ip address of the filer")
	CmdCompress.Flags().UintVar(&compressLevel, "compress_level", 3, "Compression level 1-9. 1 runs fastest with low compression rate. 9 runs slowest with high compression rate.")
}

// Invoking clp to compress the logs.
func compressLog(targetPath string, resultPath string) error {
	cmd := exec.Command(clpPath, "c", resultPath, targetPath, "--compression-level", fmt.Sprintf("%v", compressLevel))
	return cmd.Run()
}

func removeFromList(l []string, ele string) []string {
	ret := make([]string, 0, len(l)-1)
	for _, e := range l {
		if e != ele {
			ret = append(ret, e)
		}
	}
	return ret
}

func downloadDir(remotePath string, localPath string) error {
	// TODO
	return nil
}

// Wrapper for weed.UploadFile for go routine
func uploadVolume(volumeAddr string, fid string, path string, wg *sync.WaitGroup) {
	err := weed.UploadFile(volumeAddr, fid, path)
	if err != nil {
		log.Fatalf("Upload file %v to volume %v fid %v fails.", path, volumeAddr, fid)
	}
	wg.Done()
}

var archiveFiles = []string{"logypte_dict", "logtype_segindex", "metadata", "metadata_db", "var_dict", "var_segindex"}

func uploadArchive(archiveDir string, index int, fids []string, numSegments []int, wg *sync.WaitGroup) {
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
	numFiles := 6 + len(segments)
	numSegments[index] = len(segments)
	// Get a NeedleID from master
	key, err := weed.AssignFileKey(MasterAddr, numFiles)
	if err != nil {
		log.Fatalln("Get new NeedleID fails.")
	}
	fids[index] = key.Fid
	// Upload files
	for i, filename := range archiveFiles {
		wg.Add(1)
		go uploadVolume(key.PublicUrl, fmt.Sprintf("%v_%v", key.Fid, i), filepath.Join(archiveDir, filename), wg)
	}
	// Upload segments
	for i, segname := range segments {
		wg.Add(1)
		go uploadVolume(key.PublicUrl, fmt.Sprintf("%v_%v", key.Fid, 6+i), filepath.Join(archiveDir, "s", segname), wg)
	}

	wg.Done()
}

func getArchive(archives []metadata.ArchiveMetadata, archiveID string) *metadata.ArchiveMetadata {
	for i, _ := range archives {
		if archives[i].ArchiveID == archiveID {
			return &archives[i]
		}
	}
	return nil
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

	// Generate metadata
	archiveMetadatas, fileMetadatas, err := GetMetadata(filepath.Join(compressedDir, "metadata.db"))
	if err != nil {
		return
	}
	for i, _ := range fileMetadatas {
		fileMetadatas[i].Tag = tag
	}

	// Upload archives to volume servers
	archiveDir, err := os.Open(compressedDir)
	if err != nil {
		log.Println("Open compressed directory fails.")
		return
	}
	defer archiveDir.Close()
	archives, err := archiveDir.Readdirnames(0)
	if err != nil {
		log.Println("Walk compressed directory fails.")
		return
	}
	archives = removeFromList(archives, "metadata.db")
	wg := sync.WaitGroup{}
	fids := make([]string, len(archives))
	numSegments := make([]int, len(archives))
	for i, archive := range archives {
		wg.Add(1)
		go uploadArchive(filepath.Join(compressedDir, archive), i, fids, numSegments, &wg)
	}
	wg.Wait()

	// Update metadata with fids and number of segments
	for i, archiveID := range archives {
		archive := getArchive(archiveMetadatas, archiveID)
		archive.Fid = fids[i]
		archive.NumSegments = numSegments[i]
	}

	// Upload metadata
	err = MetadataService.AddMetadata(archiveMetadatas, fileMetadatas)
	if err != nil {
		return
	}
}

var archiveQuery = "SELECT id, uncompressed_size, size FROM archives"
var fileQuery = "SELECT path, begin_timestamp, end_timestamp, num_uncompressed_bytes, num_messages, archive_id FROM files"

func GetMetadata(dbFile string) ([]metadata.ArchiveMetadata, []metadata.FileMetadata, error) {
	conn, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Println("Open metadata.db fails.", err)
		return nil, nil, err
	}
	defer conn.Close()
	err = conn.Ping()
	if err != nil {
		log.Println("Open metadata.db fails.", err)
		return nil, nil, err
	}
	// Get archive metadata
	archives := make([]metadata.ArchiveMetadata, 0)
	// Map from archive_id in db to index in archives array.
	archiveRows, err := conn.Query(archiveQuery)
	if err != nil {
		log.Println("Get archive metadata fails.", err)
		return nil, nil, err
	}
	for archiveRows.Next() {
		var id string
		var uncompressedSize, size uint64
		err = archiveRows.Scan(&id, &uncompressedSize, &size)
		if err != nil {
			log.Println("Get archive metadata fails.", err)
			archiveRows.Close()
			return nil, nil, err
		}
		archives = append(archives, metadata.ArchiveMetadata{
			UncompressedSize: uncompressedSize,
			Size:             size,
			ArchiveID:        id,
		})
	}
	archiveRows.Close()
	// Get file metadata
	files := make([]metadata.FileMetadata, 0)
	fileRows, err := conn.Query(fileQuery)
	if err != nil {
		log.Println("Get file metadata fails.", err)
		return nil, nil, err
	}
	for fileRows.Next() {
		var archiveID, path string
		var beginTimestamp, endTimestamp, uncompressedBytes, numMessages uint64
		err = fileRows.Scan(&path, &beginTimestamp, &endTimestamp, &uncompressedBytes, numMessages, &archiveID)
		if err != nil {
			log.Println("Get file metadata fails.", err)
			fileRows.Close()
			return nil, nil, err
		}
		files = append(files, metadata.FileMetadata{
			FilePath:          path,
			Tag:               "",
			BeginTimestamp:    beginTimestamp,
			EndTimestamp:      endTimestamp,
			UncompressedBytes: uncompressedBytes,
			NumMessages:       numMessages,
			ArchiveID:         archiveID,
		})
	}
	fileRows.Close()
	return archives, files, nil
}
