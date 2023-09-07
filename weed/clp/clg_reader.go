package clp

type ClgFileInfo struct {
	Offset uint64
	Size   uint32
}

type ClgFileTable int

const (
	CLG_FILE_TABLE_LOGTYPE_DICT ClgFileTable = iota
	CLG_FILE_TABLE_LOGTYPE_SEGINDEX
	CLG_FILE_TABLE_METADATA
	CLG_FILE_TABLE_METADATA_DB
	CLG_FILE_TABLE_VAR_DICT
	CLG_FILE_TABLE_VAR_SEGINDEX
)

var CLG_file_name = []string{
	"logtype.dict",
	"logtype.segindex",
	"metadata",
	"metadata.db",
	"var.dict",
	"var.segindex",
}

type ClgFiles struct {
	Path     string
	Files    [6]ClgFileInfo
	Nseg     int
	Segments []ClgFileInfo
}

// func CallClgSearch() (err error) {
//Get the json body first
// var options map[string]string
// err = json.NewDecoder(r.Body).Decode(&options)
// if err != nil {
// 	glog.V(2).Infof("decoding json %s: %v", r.URL.Path, err)
// 	w.WriteHeader(http.StatusBadRequest)
// 	return
// }

// //collect all the archive files
// nSegsS, exist := options["num_segs"]
// if !exist {
// 	err = errors.New("Missing option: num-segs")
// 	glog.V(2).Infof("decoding json %s: %v", r.URL.Path, err)
// 	w.WriteHeader(http.StatusBadRequest)
// 	return
// }
// nSegs, err := strconv.Atoi(nSegsS)

// //now read all the files
// cookie := n.Cookie

// readOption := &storage.ReadOption{
// 	ReadDeleted:    r.FormValue("readDeleted") == "true",
// 	HasSlowRead:    false,
// 	ReadBufferSize: 50 * 1024 * 1024,
// }

// readOption.AttemptMetaOnly, readOption.MustMetaOnly = false, false

// var count int
// count, err = store.ReadVolumeNeedle(*vid, n, readOption, nil)

//fill the struct
//Note that seg_array is C array

//call CLG using API
// 	return

// }
