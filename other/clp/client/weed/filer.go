package weed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type File struct {
	Id   string `json:"fid"`
	Name string `json:"name"`
}

type Directory struct {
	Path    string  `json:"Directory"`
	Files   []*File `json:"Files"`
	SubDirs []*File `json:"SubDirectories"`
}

func ListDir(filerAddr string, path string) (*Directory, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%v/%v", filerAddr, path), nil)
	if err != nil {
		log.Println("Create get directory request fails.", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Send get directory request fails.", err)
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	dirResp := new(Directory)
	if err = decoder.Decode(dirResp); err != nil {
		log.Println("Decode get directory result fails.", err)
		return nil, err
	}
	return dirResp, nil
}

func DownloadFile(filerAddr string, path string, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		log.Println("Open file fails.", err)
		return err
	}
	defer file.Close()

	resp, err := http.Get(fmt.Sprintf("http://%v/%v", filerAddr, path))
	if err != nil {
		log.Println("Get file fails.", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Get file bad status: ", resp.Status)
		return errors.New(fmt.Sprintf("Donwload file bad status: %v", resp.Status))
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println("Copy to file fails.", err)
	}
	return err
}
