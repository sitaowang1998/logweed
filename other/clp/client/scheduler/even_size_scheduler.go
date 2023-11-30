package scheduler

import (
	"clp_client/metadata"
	"clp_client/weed"
	"sort"
)

// EvenSizeScheduler Schedule the archives to be searched evenly based on the size of archive
type EvenSizeScheduler struct {
	archives []ArchiveInfo
}

func newEvenSizeScheduler(archives []ArchiveInfo) EvenSizeScheduler {
	return EvenSizeScheduler{archives: archives}
}

func (s EvenSizeScheduler) Schedule() SchedulePlan {
	if len(s.archives) == 0 {
		return nil
	}
	plan := make(SchedulePlan)

	sort.Slice(s.archives, func(i, j int) bool {
		return s.archives[i].size < s.archives[j].size
	})

	for _, archive := range s.archives {
		for _, ip := range archive.ips {
			_, ok := plan[ip]
			if !ok {
				plan[ip] = nil
			}
		}
	}

	// Pre-allocate array for plan
	expectedLength := len(s.archives) / len(plan)
	for ip, _ := range plan {
		plan[ip] = make([]metadata.ArchiveMetadata, 0, expectedLength)
	}

	// Fill the largest remaining archive into the emptiest server
	sizes := make(map[weed.VolumeAddr]uint64)
	for ip, _ := range plan {
		sizes[ip] = 0
	}
	for i := len(s.archives) - 1; i >= 0; i-- {
		ip := getMinSizeIp(sizes, s.archives[i].ips)
		plan[ip] = append(plan[ip], s.archives[i].getMetadata())
		sizes[ip] += s.archives[i].size
	}

	return plan
}

func getMinSizeIp(sizes map[weed.VolumeAddr]uint64, ips []weed.VolumeAddr) weed.VolumeAddr {
	candidateSizes := make([]uint64, 0, len(ips))
	for _, ip := range ips {
		candidateSizes = append(candidateSizes, sizes[ip])
	}
	minIndex := 0
	for i := 1; i < len(candidateSizes); i++ {
		if candidateSizes[i] <= candidateSizes[minIndex] {
			minIndex = i
		}
	}
	return ips[minIndex]
}
