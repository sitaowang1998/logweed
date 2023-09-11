package metadata

type ArchiveMetadata struct {
	UncompressedSize uint64
	Size             uint64
	Fid              string
	NumSegments      int
	ArchiveID        string
}

type FileMetadata struct {
	FilePath          string
	Tag               string
	BeginTimestamp    int64
	EndTimestamp      int64
	UncompressedBytes uint64
	NumMessages       uint64
	ArchiveID         string
}

type MetadataService interface {
	Connect(usr string, pwd string, addr string) error
	Close() error
	InitService() error
	ListTags() ([]string, error)
	GetFiles(tag string) ([]string, error)
	AddMetadata(archives []ArchiveMetadata, files []FileMetadata) error
	Search(tag string, beginTimestamp uint64, endTimestamp uint64) ([]ArchiveMetadata, error)
}
