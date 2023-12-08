package weed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type VolumeAddr struct {
	PublicUrl string `json:"publicUrl"`
	Url       string `json:"url"`
}

type volumeLookupResp struct {
	Locations []VolumeAddr `json:"locations"`
}

func LookupVolume(masterAddr string, vid string) ([]VolumeAddr, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v/dir/lookup?volumeId=%v", masterAddr, vid))
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
	return v.Locations, nil
}

type FileKey struct {
	Count     int    `json:"count"`
	Fid       string `json:"fid"`
	Url       string `json:"url"`
	PublicUrl string `json:"publicUrl"`
}

func AssignFileKey(masterAddr string, count int) (FileKey, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v/dir/assign?count=%v", masterAddr, count))
	if err != nil {
		log.Println("Assign file key fails.", err)
		return FileKey{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Read assign file key response fails.", err)
		return FileKey{}, err
	}
	var k FileKey
	err = json.Unmarshal(body, &k)
	if err != nil {
		log.Println("Parse assigned file key fails.", err)
		return FileKey{}, err
	}
	if k.Count != count {
		log.Printf("Assigned file key count not match. Get %v, expect %v.\n", k.Count, count)
		return k, errors.New("file count not match")
	}
	return k, nil
}
