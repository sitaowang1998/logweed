package cmd

import (
	"clp_client/metadata"
	"clp_client/scheduler"
	"clp_client/weed"
	"fmt"
	"github.com/spf13/cobra"
	"log"
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

func searchArchiveInVolume(archive metadata.ArchiveMetadata, ip string, query string, bts uint64, ets uint64, results *[]string, mutex *sync.Mutex, wg *sync.WaitGroup) {
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

	log.Printf("Volume search time: %d ms.\n",
		volumeTime.Sub(startTime).Milliseconds(),
	)

	mutex.Lock()
	*results = append(*results, result...)
	mutex.Unlock()
}

func getVolumeAddress(volumeId string, addresses map[weed.VolumeAddr]struct{}, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	ips, err := weed.LookupVolume(MasterAddr, volumeId)
	if err != nil {
		log.Printf("Error getting volume address for %s", volumeId)
		log.Println(err)
		os.Exit(1)
	}

	mutex.Lock()
	defer mutex.Unlock()
	for _, ip := range ips {
		addresses[ip] = struct{}{}
	}
}

func getVolumeAddresses(archives []metadata.ArchiveMetadata) []weed.VolumeAddr {
	var wg sync.WaitGroup
	var mutex sync.Mutex

	vids := make(map[string]struct{})

	for _, archive := range archives {
		vid := weed.ExtractVolumeId(archive.Fid)
		vids[vid] = struct{}{}
	}

	addresses := make(map[weed.VolumeAddr]struct{})
	for vid := range vids {
		wg.Add(1)
		getVolumeAddress(vid, addresses, &mutex, &wg)
	}

	wg.Wait()

	ips := make([]weed.VolumeAddr, len(addresses))
	i := 0
	for addr := range addresses {
		ips[i] = addr
		i++
	}

	return ips
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

	// Get volume server addresses
	volumeIps := getVolumeAddresses(archives)

	// Generate schedule plan
	archiveInfo := make([]scheduler.ArchiveInfo, 0, len(archives))
	for _, archive := range archives {
		archiveInfo = append(archiveInfo, scheduler.NewArchiveInfo(
			archive, volumeIps))
	}
	sched := scheduler.NewEvenSizeScheduler(archiveInfo)
	schedulePlan := sched.Schedule()

	// Search in parallel
	results := make([]string, 0)
	mutex := sync.Mutex{}
	var wg sync.WaitGroup
	for ip, archives := range schedulePlan {
		for _, archive := range archives {
			wg.Add(1)
			go searchArchiveInVolume(archive, ip.PublicUrl, query, bts, ets, &results, &mutex, &wg)
		}
	}
	wg.Wait()
	log.Println("Search complete.")

	for _, line := range results {
		fmt.Println(line)
	}
}
