package metadata

type ArchiveMetadata struct {
	UncompressedSize uint64
	Size             uint64
	Fid              string
	NumSegments      int
}

type FileMetadata struct {
	FilePath          string
	Tag               string
	BeginTimestamp    uint64
	EndTimestamp      uint64
	UncompressedBytes uint64
	NumMessages       uint64
	ArchiveID         int64
}

type MetadataService interface {
	Connect(usr string, pwd string, addr string) error
	Close() error
	InitService() error
	ListTags() ([]string, error)
	GetFiles(tag string) ([]string, error)
	// AddMetadata Insert metadata into database. All insertions are executed in a transaction.
	// ArchiveID in ArchiveMetadata is not used. ArchiveID in FileMetadata in the index into archives,
	// not the archive_id in archives table in database, and is assumed to be within range.
	AddMetadata(archives []ArchiveMetadata, files []FileMetadata) error
	Search(tag string, beginTimestamp uint64, endTimestamp uint64) ([]ArchiveMetadata, error)
}
