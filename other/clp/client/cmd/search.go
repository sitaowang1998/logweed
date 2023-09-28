package cmd

import (
	"clp_client/metadata"
	"clp_client/weed"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	btsStr string
	etsStr string
)

var CmdSearch = &cobra.Command{
	Use:   "search <tag> <query> <--bts bts> <--ets ets>",
	Short: "Search the logs in SeaweedFS.",
	Long:  "Search the log files associated with the tag.",
	Args:  cobra.ExactArgs(2),
	Run:   search,
}

func init() {
	CmdSearch.Flags().StringVar(&btsStr, "bts", "", "begin timestamp")
	CmdSearch.Flags().StringVar(&etsStr, "ets", "", "end timestamp")
	CmdSearch.MarkFlagRequired("bts")
	CmdSearch.MarkFlagRequired("ets")
}

func parseTimeStamp(ts string) (uint64, error) {
	// First try parse it as int
	res, err := strconv.ParseUint(ts, 10, 64)
	if err == nil {
		return res, nil
	}
	// Try parse as timestamp string in several formats
	t, err := time.Parse(time.RFC822, ts)
	if err == nil {
		return uint64(t.UnixMilli()), nil
	}
	t, err = time.Parse(time.RFC850, ts)
	if err == nil {
		return uint64(t.UnixMilli()), nil
	}
	t, err = time.Parse(time.RFC1123, ts)
	if err == nil {
		return uint64(t.UnixMilli()), nil
	}
	t, err = time.Parse(time.RFC3339, ts)
	if err == nil {
		return uint64(t.UnixMilli()), nil
	}
	// Parse fail
	return 0, nil
}

func searchArchiveInVolume(archive *metadata.ArchiveMetadata, ip string, query string, bts uint64, ets uint64, results *[]string, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()

	var args []string
	args = append(args, query)
	// For now, remove the timestamp filter for clg
	// args = append(args, "--tge")
	// args = append(args, strconv.FormatUint(bts, 10))
	// args = append(args, "--tle")
	// args = append(args, strconv.FormatUint(ets, 10))
	result, err := weed.ClgRemoteSearch(ip, weed.ClgSearchRequest{
		Fid:              archive.Fid,
		NumSegments:      uint64(archive.NumSegments),
		ArchiveID:        archive.ArchiveID,
		UncompressedSize: archive.UncompressedSize,
		Size:             archive.Size,
		Args:             args,
	})
	if err != nil {
		os.Exit(1)
	}

	volumeTime := time.Now()

	log.Printf("Total search: %d ms. Master: %d ms. Volume: %d ms.\n",
		volumeTime.Sub(startTime).Milliseconds(),
		0,
		volumeTime.Sub(startTime).Milliseconds(),
	)

	mutex.Lock()
	*results = append(*results, result...)
	mutex.Unlock()
}

func search(cmd *cobra.Command, args []string) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	// Parse the arguments and flags
	tag := args[0]
	query := args[1]
	if tag == "" {
		log.Println("Tag must not be empty.")
		return
	}
	if query == "" {
		log.Println("Query must not be empty.")
		return
	}
	bts, err := parseTimeStamp(btsStr)
	if err != nil {
		log.Println("Invalid begin timestamp.")
		return
	}
	ets, err := parseTimeStamp(etsStr)
	if err != nil {
		log.Println("Invalid end timestamp.")
		return
	}

	connectMetadataServer()

	log.Println("Connected to metadata db.")
	// Request archives from metadata service
	archives, err := MetadataService.Search(tag, bts, ets)
	if err != nil {
		return
	}
	log.Printf("Need to search %d archives.\n", len(archives))

	// Shuffle the archives and assign ips to archives
	numArchives := len(archives)
	rand.Shuffle(numArchives, func(i, j int) {
		archives[i], archives[j] = archives[j], archives[i]
	})

	ips := []string{"10.1.0.7",
		"10.1.0.10",
		"10.1.0.12",
		"10.1.0.15",
		"10.1.0.16",
		"10.1.0.17",
		"10.1.0.18",
		"10.1.0.19",
	}

	archivePerIp := numArchives / len(ips)
	if numArchives%len(ips) > 0 {
		archivePerIp += 1
	}

	// Search in parallel
	results := make([]string, 0)
	mutex := sync.Mutex{}
	var wg sync.WaitGroup
	for i := 0; i < numArchives; i++ {
		wg.Add(1)
		ip := ips[i/archivePerIp]
		go searchArchiveInVolume(&archives[i], ip, query, bts, ets, &results, &mutex, &wg)
	}
	wg.Wait()
	log.Println("Search complete.")

	for _, line := range results {
		fmt.Println(line)
	}
}
