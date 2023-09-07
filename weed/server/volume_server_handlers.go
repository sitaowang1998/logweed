package weed_server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/seaweedfs/seaweedfs/weed/storage/backend"
	"github.com/seaweedfs/seaweedfs/weed/storage/needle"
	"github.com/seaweedfs/seaweedfs/weed/storage/types"
	"github.com/seaweedfs/seaweedfs/weed/util"

	"github.com/seaweedfs/seaweedfs/weed/glog"
	"github.com/seaweedfs/seaweedfs/weed/security"
	"github.com/seaweedfs/seaweedfs/weed/stats"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seaweedfs/seaweedfs/weed/clp"
)

/*

If volume server is started with a separated public port, the public port will
be more "secure".

Public port currently only supports reads.

Later writes on public port can have one of the 3
security settings:
1. not secured
2. secured by white list
3. secured by JWT(Json Web Token)

*/

func (vs *VolumeServer) privateStoreHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "SeaweedFS Volume "+util.VERSION)
	if r.Header.Get("Origin") != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	stats.VolumeServerRequestCounter.WithLabelValues(r.Method).Inc()
	start := time.Now()
	defer func(start time.Time) {
		stats.VolumeServerRequestHistogram.WithLabelValues(r.Method).Observe(time.Since(start).Seconds())
	}(start)
	switch r.Method {
	case "GET", "HEAD":
		stats.ReadRequest()
		vs.inFlightDownloadDataLimitCond.L.Lock()
		inFlightDownloadSize := atomic.LoadInt64(&vs.inFlightDownloadDataSize)
		for vs.concurrentDownloadLimit != 0 && inFlightDownloadSize > vs.concurrentDownloadLimit {
			select {
			case <-r.Context().Done():
				glog.V(4).Infof("request cancelled from %s: %v", r.RemoteAddr, r.Context().Err())
				w.WriteHeader(http.StatusInternalServerError)
				vs.inFlightDownloadDataLimitCond.L.Unlock()
				return
			default:
				glog.V(4).Infof("wait because inflight download data %d > %d", inFlightDownloadSize, vs.concurrentDownloadLimit)
				vs.inFlightDownloadDataLimitCond.Wait()
			}
			inFlightDownloadSize = atomic.LoadInt64(&vs.inFlightDownloadDataSize)
		}
		vs.inFlightDownloadDataLimitCond.L.Unlock()
		vs.GetOrHeadHandler(w, r)
	case "DELETE":
		stats.DeleteRequest()
		vs.guard.WhiteList(vs.DeleteHandler)(w, r)
	case "PUT", "POST":
		contentLength := getContentLength(r)
		// exclude the replication from the concurrentUploadLimitMB
		if r.URL.Query().Get("type") != "replicate" && vs.concurrentUploadLimit != 0 {
			startTime := time.Now()
			vs.inFlightUploadDataLimitCond.L.Lock()
			inFlightUploadDataSize := atomic.LoadInt64(&vs.inFlightUploadDataSize)
			for inFlightUploadDataSize > vs.concurrentUploadLimit {
				//wait timeout check
				if startTime.Add(vs.inflightUploadDataTimeout).Before(time.Now()) {
					vs.inFlightUploadDataLimitCond.L.Unlock()
					err := fmt.Errorf("reject because inflight upload data %d > %d, and wait timeout", inFlightUploadDataSize, vs.concurrentUploadLimit)
					glog.V(1).Infof("too many requests: %v", err)
					writeJsonError(w, r, http.StatusTooManyRequests, err)
					return
				}
				glog.V(4).Infof("wait because inflight upload data %d > %d", inFlightUploadDataSize, vs.concurrentUploadLimit)
				vs.inFlightUploadDataLimitCond.Wait()
				inFlightUploadDataSize = atomic.LoadInt64(&vs.inFlightUploadDataSize)
			}
			vs.inFlightUploadDataLimitCond.L.Unlock()
		}
		atomic.AddInt64(&vs.inFlightUploadDataSize, contentLength)
		defer func() {
			atomic.AddInt64(&vs.inFlightUploadDataSize, -contentLength)
			if vs.concurrentUploadLimit != 0 {
				vs.inFlightUploadDataLimitCond.Signal()
			}
		}()

		// processs uploads
		stats.WriteRequest()
		vs.guard.WhiteList(vs.PostHandler)(w, r)

	case "OPTIONS":
		stats.ReadRequest()
		w.Header().Add("Access-Control-Allow-Methods", "PUT, POST, GET, DELETE, OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "*")
	}
}

func getContentLength(r *http.Request) int64 {
	contentLength := r.Header.Get("Content-Length")
	if contentLength != "" {
		length, err := strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return 0
		}
		return length
	}
	return 0
}

type ClgSearchRequest struct {
	Fid              string   `json:"fid"`
	NumSegments      uint64   `json:"nseg"`
	ArchiveID        string   `json:"archid"`
	UncompressedSize uint64   `json:"uncompressed_size"`
	Size             uint64   `json:"size"`
	Args             []string `json:"args"`
}

const dropArchiveQuery = `DROP TABLE IF EXISTS archives;`

const createArchiveQuery = `
CREATE TABLE archives (
    id TEXT PRIMARY KEY,
    uncompressed_size INTEGER,
    size INTEGER,
    creator_id TEXT,
    creation_ix INTEGER
) WITHOUT ROWID;
`

const createArchiveIndexQuery = `
CREATE INDEX archives_creation_order ON archives (creator_id,creation_ix);
`

const insertArchiveQuery = `
INSERT INTO archives(id, uncompressed_size, size, creator_id, creation_ix) VALUES(
	?, ?, ?, ?, ?
);
`

func createCLGDB(dir string, archiveID string, uncompressedSize uint64, size uint64) error {
	filePath := dir + "/metadata.db"
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(dropArchiveQuery); err != nil {
		glog.V(0).Infoln("Drop table fail", err)
		return err
	}
	if _, err := db.Exec(createArchiveQuery); err != nil {
		glog.V(0).Infoln("Create table fail", err)
		return err
	}
	if _, err := db.Exec(createArchiveIndexQuery); err != nil {
		glog.V(0).Infoln("Create index fail", err)
	}
	if _, err := db.Exec(insertArchiveQuery, archiveID, uncompressedSize, size, "", 0); err != nil {
		glog.V(0).Infoln("Insert table fail", err)
		return err
	}
	return nil
}

func (vs *VolumeServer) clgHandler(w http.ResponseWriter, r *http.Request) {
	n := new(needle.Needle)
	w.Header().Set("Server", "SeaweedFS Volume "+util.VERSION)
	if r.Header.Get("Origin") != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	glog.V(0).Infoln("Clg: Receive Clg request.")
	// Decode json request
	decoder := json.NewDecoder(r.Body)
	var request ClgSearchRequest
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find the clg files
	var path string
	path = request.Fid
	archId := request.ArchiveID
	glog.V(0).Infof("Got fid: %s", path)
	// Parse fid
	commaIndex := strings.LastIndex(path, ",")
	vid, err := strconv.ParseUint(path[:commaIndex], 10, 64)
	fid := path[commaIndex+1:]
	err = n.ParsePath(fid)
	if err != nil {
		glog.V(2).Infof("parsing fid %s: %volume", r.URL.Path, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	numSegments := request.NumSegments
	volume := vs.store.GetVolume(needle.VolumeId(vid))
	nm := volume.GetNm()

	diskFile, ok := volume.DataBackend.(*backend.DiskFile)
	if !ok {
		panic("not disk file")
	}

	// Create the clg directory
	archPath := "/mnt/ramdisk/archives/" + archId + "/" + archId
	err = os.MkdirAll(archPath, 0777)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	glog.V(0).Infof("fid: %x dir: %s", fid, archPath)
	// for each archive data, get the offset and size
	var clgfiles clp.ClgFiles
	for i := uint64(0); i < 6; i++ {
		nv, ok := nm.Get(types.NeedleId(uint64(n.Id) + i))
		if !ok || nv.Offset.IsZero() {
			glog.V(0).Infoln("Failed get fid", nv)
			panic(err)
			return
		}
		glog.V(0).Infoln("Got needle: ", nv)

		readOffset := nv.Offset.ToActualOffset() + types.NeedleHeaderSize
		// read the size
		buf := make([]byte, 4)
		_, err = diskFile.File.ReadAt(buf, int64(readOffset))
		if err != nil {
			panic(err)
		}
		readSize := util.BytesToUint32(buf)
		readOffset += 4
		clgfiles.Files[i].Offset = uint64(readOffset)
		clgfiles.Files[i].Size = uint32(readSize)
		if clgfiles.Files[i].Size < 1024 {
			clgfiles.Files[i].Size += 4
		}
		target, err := os.Create(archPath + "/" + clp.CLG_file_name[i])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer target.Close()
		glog.V(0).Infof("created file: %s, offset: %d, size: %d",
			archPath+"/"+clp.CLG_file_name[i], clgfiles.Files[i].Offset,
			clgfiles.Files[i].Size)
		// Call copy_file_range
		// glog.V(0).Infof("Copy file range: From: %d, To: %d, offset: %d, size: %d",
		// 	fd, target.Fd(),
		// 	clgfiles.Files[i].Offset, clgfiles.Files[i].Size)
		// _, _, errno := unix.Syscall6(
		// 	unix.SYS_COPY_FILE_RANGE,
		// 	uintptr(fd), // Source file descriptor
		// 	uintptr(unsafe.Pointer(&(clgfiles.Files[i].Offset))), // Source file offset
		// 	uintptr(target.Fd()),            // Destination file descriptor
		// 	0,                               // Destination file offset (0 for appending)
		// 	uintptr(clgfiles.Files[i].Size), // Number of bytes to copy
		// 	0,                               // Copy flags (0 for default)
		// )
		// if errno != 0 {
		// 	err = errno
		// 	glog.V(0).Infof("Error: %s", err.Error())
		// }

		// Alternative: use io.CopyN
		diskFile.File.Seek(int64(clgfiles.Files[i].Offset), 0)
		copied, err := io.CopyN(target, diskFile.File, int64(clgfiles.Files[i].Size))
		if err != nil {
			glog.V(0).Infof("Error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if copied != int64(clgfiles.Files[i].Size) {
			glog.V(0).Infof("Error: copied %d bytes, expected %d bytes", copied, clgfiles.Files[i].Size)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = os.MkdirAll(archPath+"/s/", 0777)
	if err != nil {
		glog.V(0).Infof("Error: cannot create directory %s", archPath+"/s/")
		w.WriteHeader(http.StatusInternalServerError)
	}

	// for each archive segment, get the offset and size
	for i := uint64(6); i < 6+numSegments; i++ {
		nv, ok := nm.Get(types.NeedleId(uint64(n.Id) + i))
		if !ok || nv.Offset.IsZero() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		readOffset := nv.Offset.ToActualOffset() + types.NeedleHeaderSize
		// read the size
		buf := make([]byte, 4)
		_, err = diskFile.File.ReadAt(buf, int64(readOffset))
		if err != nil {
			panic(err)
		}
		readSize := util.BytesToUint32(buf)
		readOffset += 4
		var seg clp.ClgFileInfo
		seg.Offset = uint64(readOffset)
		seg.Size = uint32(readSize)
		if seg.Size < 1024 {
			seg.Size += 4
		}
		clgfiles.Segments = append(clgfiles.Segments, seg)

		target, err := os.Create(archPath + "/s/" + strconv.FormatUint(i-6, 10))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer target.Close()
		glog.V(0).Infof("created file: %s, offset: %d, size: %d",
			archPath+"/s/"+strconv.FormatUint(i-6, 10), seg.Offset,
			seg.Size)
		// Call copy_file_range

		// Alternative: use io.CopyN
		diskFile.File.Seek(int64(seg.Offset), 0)
		copied, err := io.CopyN(target, diskFile.File, int64(seg.Size))
		if err != nil {
			glog.V(0).Infof("Error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if copied != int64(seg.Size) {
			glog.V(0).Infof("Error: copied %d bytes, expected %d bytes", copied, seg.Size)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Use copy_file_range syscall
		// _, _, errno := unix.Syscall6(
		// 	unix.SYS_COPY_FILE_RANGE,
		// 	uintptr(fd),                            // Source file descriptor
		// 	uintptr(unsafe.Pointer(&(seg.Offset))), // Source file offset
		// 	uintptr(target.Fd()),                   // Destination file descriptor
		// 	0,                                      // Destination file offset (0 for appending)
		// 	uintptr(seg.Size),                      // Number of bytes to copy
		// 	0,                                      // Copy flags (0 for default)
		// )
		// if errno != 0 {
		// 	err = errno
		// 	glog.V(0).Infof("Error: %s", err.Error())
		// }
	}
	glog.V(0).Infoln("Clg: Copy file completes.")

	// Create dummy db file
	err = createCLGDB("/mnt/ramdisk/archives/"+archId, archId, request.UncompressedSize, request.Size)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	glog.V(0).Infoln("Clg: Create dummy db completes.")

	// Spawn the clg process
	clg_bin := "/home/sitao/clp/bin/clg"
	var args []string
	args = append(args, "/mnt/ramdisk/archives/"+archId)
	args = append(args, request.Args...)

	cmd := exec.Command(clg_bin, args...)

	output, err := cmd.Output()
	if err != nil {
		glog.V(0).Infoln("Clg search fail", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	glog.V(0).Infoln("Clg: Search completes.")

	// Send the output back to the client
	w.Write(output)
	glog.V(0).Infoln("Clg: Send reply completes.")
}

func (vs *VolumeServer) publicReadOnlyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "SeaweedFS Volume "+util.VERSION)
	if r.Header.Get("Origin") != "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	switch r.Method {
	case "GET", "HEAD":
		stats.ReadRequest()
		vs.inFlightDownloadDataLimitCond.L.Lock()
		inFlightDownloadSize := atomic.LoadInt64(&vs.inFlightDownloadDataSize)
		for vs.concurrentDownloadLimit != 0 && inFlightDownloadSize > vs.concurrentDownloadLimit {
			glog.V(4).Infof("wait because inflight download data %d > %d", inFlightDownloadSize, vs.concurrentDownloadLimit)
			vs.inFlightDownloadDataLimitCond.Wait()
			inFlightDownloadSize = atomic.LoadInt64(&vs.inFlightDownloadDataSize)
		}
		vs.inFlightDownloadDataLimitCond.L.Unlock()
		vs.GetOrHeadHandler(w, r)
	case "OPTIONS":
		stats.ReadRequest()
		w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "*")
	}
}

func (vs *VolumeServer) maybeCheckJwtAuthorization(r *http.Request, vid, fid string, isWrite bool) bool {

	var signingKey security.SigningKey

	if isWrite {
		if len(vs.guard.SigningKey) == 0 {
			return true
		} else {
			signingKey = vs.guard.SigningKey
		}
	} else {
		if len(vs.guard.ReadSigningKey) == 0 {
			return true
		} else {
			signingKey = vs.guard.ReadSigningKey
		}
	}

	tokenStr := security.GetJwt(r)
	if tokenStr == "" {
		glog.V(1).Infof("missing jwt from %s", r.RemoteAddr)
		return false
	}

	token, err := security.DecodeJwt(signingKey, tokenStr, &security.SeaweedFileIdClaims{})
	if err != nil {
		glog.V(1).Infof("jwt verification error from %s: %v", r.RemoteAddr, err)
		return false
	}
	if !token.Valid {
		glog.V(1).Infof("jwt invalid from %s: %v", r.RemoteAddr, tokenStr)
		return false
	}

	if sc, ok := token.Claims.(*security.SeaweedFileIdClaims); ok {
		if sepIndex := strings.LastIndex(fid, "_"); sepIndex > 0 {
			fid = fid[:sepIndex]
		}
		return sc.Fid == vid+","+fid
	}
	glog.V(1).Infof("unexpected jwt from %s: %v", r.RemoteAddr, tokenStr)
	return false
}
