package weed

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type VolumeAddr struct {
	publicUrl string
	url       string
}

type volumeLookupResp struct {
	locations []VolumeAddr
}

func LookupVolume(masterAddr string, vid string) ([]VolumeAddr, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v/dir/lookup?volumeID=%v", masterAddr, vid))
	if err != nil {
		log.Println("Lookup volume fails.", err)
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var v volumeLookupResp
	if err := decoder.Decode(&v); err != nil {
		log.Println("Parse volume lookup fails.", err)
		return nil, err
	}
	return v.locations, nil
}

type FileKey struct {
	count     int
	fid       string
	url       string
	publicUrl string
}

func AssignFileKey(masterAddr string, count int) (FileKey, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v/dir/assign?count=%v", masterAddr, count))
	if err != nil {
		log.Println("Assign file key fails.", err)
		return FileKey{}, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var k FileKey
	if err := decoder.Decode(&k); err != nil {
		log.Println("Parse assigned file key fails.", err)
		return FileKey{}, err
	}
	return k, nil
}
