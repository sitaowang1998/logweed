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
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	clpPath       string
	filerAddr     string
	compressLevel uint
	uploadOnly    bool
)

var CmdCompress = &cobra.Command{
	Use:   "compress <tag> <dir> [--filer ipAddr] [--clp_path clpPath] [--upload-only]",
	Short: "Compress logs to SeaweedFS.",
	Long:  "Compress the logs in a directory locally or on SeaweedFS.",
	Args:  cobra.ExactArgs(2),
	Run:   compress,
}

func init() {
	CmdCompress.Flags().StringVar(&clpPath, "clp_path", "", "path to clp binary")
	CmdCompress.Flags().StringVar(&filerAddr, "filer", "", "ip address of the filer")
	CmdCompress.Flags().UintVar(&compressLevel, "compress_level", 3, "Compression level 1-9. 1 runs fastest with low compression rate. 9 runs slowest with high compression rate.")
	CmdCompress.Flags().BoolVarP(&uploadOnly, "upload_only", "u", false, "Upload compressed directory directly")
}

// Invoking clp to compress the logs.
func compressLog(targetPath string, resultPath string) error {
	cmd := exec.Command(clpPath, "c", resultPath, targetPath, "--compression-level", fmt.Sprintf("%v", compressLevel), "--target-dictionaries-size", "10485760")
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

func weedDownload(remotePath string, localPath string) error {
	if !strings.HasSuffix(remotePath, "/") {
		return weed.DownloadFile(filerAddr, remotePath, localPath)
	}
	dir, err := weed.ListDir(filerAddr, remotePath)
	if err != nil {
		return err
	}
	for _, file := range dir.Files {
		err = weed.DownloadFile(filerAddr, fmt.Sprintf("%v/%v", dir.Path, file.Name), filepath.Join(localPath, file.Name))
		if err != nil {
			return err
		}
	}
	for _, subDir := range dir.SubDirs {
		err = weedDownload(fmt.Sprintf("%v/%v/", dir.Path, subDir.Name), filepath.Join(localPath, subDir.Name))
		if err != nil {
			return err
		}
	}
	return nil
}

// Wrapper for weed.UploadFile for go routine
func uploadVolume(volumeAddr string, fid string, path string) {
	err := weed.UploadFile(volumeAddr, fid, path)
	if err != nil {
		log.Fatalf("Upload file %v to volume %v fid %v fails.", path, volumeAddr, fid)
	}
}

var archiveFiles = []string{"logtype.dict", "logtype.segindex", "metadata", "metadata.db", "var.dict", "var.segindex"}

type archiveIndex struct {
	ArchiveDir string
	Index      int
}

type waitQueue struct {
	archives map[string][]archiveIndex
	inUse    map[string]bool
	mutex    sync.Mutex
}

func (w *waitQueue) addArchive(archiveDir string, index int, key weed.FileKey) {
	vid := weed.ExtractVolumeId(key.Fid)
	w.mutex.Lock()
	defer w.mutex.Unlock()
	_, ok := w.archives[vid]
	if !ok {
		w.archives[vid] = make([]archiveIndex, 0)
		w.inUse[vid] = false
	}
	w.archives[vid] = append(w.archives[vid], archiveIndex{
		ArchiveDir: archiveDir,
		Index:      index,
	})
}

func (w *waitQueue) getTask(prevVid string) *archiveIndex {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if len(prevVid) != 0 {
		w.inUse[prevVid] = false
	}
	for vid := range w.archives {
		if !w.inUse[vid] && len(w.archives[vid]) != 0 {
			archive := w.archives[vid][0]
			w.archives[vid] = w.archives[vid][1:]
			w.inUse[vid] = true
			return &archive
		}
	}
	return nil
}

func getNeedleId(archiveDir string, index int, keys []weed.FileKey, numSegments []int, queue *waitQueue, wg *sync.WaitGroup) {
	// Get number of files
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
	keys[index] = key

	queue.addArchive(archiveDir, index, key)

	wg.Done()
}

func uploadArchiveWorker(queue *waitQueue, keys []weed.FileKey, numSegments []int, wg *sync.WaitGroup) {
	defer wg.Done()
	archive := queue.getTask("")
	fid := keys[archive.Index].Fid
	vid := weed.ExtractVolumeId(fid)
	uploadArchive(archive.ArchiveDir, keys[archive.Index], numSegments[archive.Index])
	for {
		archive = queue.getTask(vid)
		if archive == nil {
			return
		}
		fid = keys[archive.Index].Fid
		vid = weed.ExtractVolumeId(fid)
		uploadArchive(archive.ArchiveDir, keys[archive.Index], numSegments[archive.Index])
	}
}

func uploadArchive(archiveDir string, key weed.FileKey, numSegments int) {
	// Upload files
	for i, filename := range archiveFiles {
		go uploadVolume(key.PublicUrl,
			fmt.Sprintf("%v_%v", key.Fid, i),
			filepath.Join(archiveDir, filename))
	}
	// Upload segments
	for i := 0; i < numSegments; i++ {
		go uploadVolume(key.PublicUrl,
			fmt.Sprintf("%v_%v", key.Fid, 6+i),
			filepath.Join(archiveDir, "s", strconv.FormatInt(int64(i), 10)))
	}
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

	connectMetadataServer()
	log.Println("Connected to metadata server.")

	compressedDir := dir
	if !uploadOnly {
		// Download file from filer if necessary
		if filerAddr != "" {
			localDir := fmt.Sprintf("/tmp/logweed/uncompressed/%v", time.Now())
			err := weedDownload(dir, localDir)
			if err != nil {
				return
			}
			dir = localDir
			log.Println("Downloaded files from seaweedFS.")
		}

		// Run compression
		compressedDir = fmt.Sprintf("/tmp/logweed/compressed/%v", time.Now())
		err := compressLog(dir, compressedDir)
		if err != nil {
			log.Println("Clp compress fails.")
			return
		}
		log.Println("Clp compressed finishes.")
	}

	// Generate metadata
	archiveMetadatas, fileMetadatas, err := GetMetadata(filepath.Join(compressedDir, "metadata.db"))
	if err != nil {
		return
	}
	for i, _ := range fileMetadatas {
		fileMetadatas[i].Tag = tag
	}
	log.Println("Got metadata from db.")

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
	keys := make([]weed.FileKey, len(archives))
	numSegments := make([]int, len(archives))
	var queue waitQueue
	numWorkers := 12 // Hardcode number of workers for now
	// Get archive id and initiate tasks
	for i, archive := range archives {
		wg.Add(1)
		go getNeedleId(filepath.Join(compressedDir, archive), i, keys, numSegments, &queue, &wg)
	}

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go uploadArchiveWorker(&queue, keys, numSegments, &wg)
	}

	wg.Wait()
	log.Println("Uploaded archives.")

	// Update metadata with fids and number of segments
	for i, archiveID := range archives {
		archive := getArchive(archiveMetadatas, archiveID)
		archive.Fid = keys[i].Fid
		archive.NumSegments = numSegments[i]
	}

	// Upload metadata
	err = MetadataService.InitService()
	if err != nil {
		return
	}
	err = MetadataService.AddMetadata(archiveMetadatas, fileMetadatas)
	if err != nil {
		return
	}
	log.Println("Updated metadata.")
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
		var uncompressedBytes, numMessages uint64
		var beginTimestamp, endTimestamp int64
		err = fileRows.Scan(&path, &beginTimestamp, &endTimestamp, &uncompressedBytes, &numMessages, &archiveID)
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
