package scheduler

import (
	"clp_client/metadata"
	"clp_client/weed"
)

type SchedulePlan map[weed.VolumeAddr][]metadata.ArchiveMetadata

type Scheduler interface {
	// Schedule Provide the next scheduling plan.
	Schedule() SchedulePlan
}

type ArchiveInfo struct {
	size             uint64
	uncompressedSize uint64
	numSegments      int
	fid              string
	archiveId        string
	ips              []weed.VolumeAddr
}

func newArchiveInfo(metadata metadata.ArchiveMetadata, ips []weed.VolumeAddr) ArchiveInfo {
	return ArchiveInfo{
		size:             metadata.Size,
		uncompressedSize: metadata.UncompressedSize,
		numSegments:      metadata.NumSegments,
		fid:              metadata.Fid,
		archiveId:        metadata.ArchiveID,
		ips:              ips,
	}
}

func (a ArchiveInfo) getMetadata() metadata.ArchiveMetadata {
	return metadata.ArchiveMetadata{
		UncompressedSize: a.uncompressedSize,
		Size:             a.size,
		Fid:              a.fid,
		NumSegments:      a.numSegments,
		ArchiveID:        a.archiveId,
	}
}
