package cmd

import (
	"clp_client/metadata"
	"clp_client/weed"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"os"
	"sort"
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

func searchArchiveInVolume(archive *metadata.ArchiveMetadata, query string, bts uint64, ets uint64, results *[]string, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	// Get volume ip address from master
	ips, err := weed.LookupVolume(MasterAddr, archive.Fid)
	if err != nil {
		os.Exit(1)
	}
	// Send search request with a random ip
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	ip := ips[random.Intn(len(ips))]
	var args []string
	args = append(args, query)
	args = append(args, "--tge")
	args = append(args, strconv.FormatUint(bts, 10))
	args = append(args, "--tle")
	args = append(args, strconv.FormatUint(ets, 10))
	result, err := weed.ClgSearch(ip.PublicUrl, weed.ClgSearchRequest{
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
	mutex.Lock()
	*results = append(*results, result...)
	mutex.Unlock()
}

func search(cmd *cobra.Command, args []string) {
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

	// Request archives from metadata service
	archives, err := MetadataService.Search(tag, bts, ets)
	if err != nil {
		return
	}

	// Search in parallel
	results := make([]string, 0)
	mutex := sync.Mutex{}
	var wg sync.WaitGroup
	for i := 0; i < len(archives); i++ {
		wg.Add(1)
		go searchArchiveInVolume(&archives[i], query, bts, ets, &results, &mutex, &wg)
	}
	wg.Wait()

	// sort the results
	sort.Strings(results)

	for _, line := range results {
		fmt.Println(line)
	}
}
