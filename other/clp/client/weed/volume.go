package weed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func UploadFile(volumeAddr string, fid string, filepath string) error {
	read, write := io.Pipe()
	m := multipart.NewWriter(write)

	go func() {
		defer write.Close()
		defer m.Close()

		part, err := m.CreateFormFile("uploadfile", filepath)
		if err != nil {
			log.Println("Create multipart fails.", err)
			return
		}
		file, err := os.Open(filepath)
		if err != nil {
			log.Println("Open file fails.", err)
			return
		}
		defer file.Close()

		if _, err = io.Copy(part, file); err != nil {
			log.Println("Copy file fails.", err)
			return
		}
	}()

	_, err := http.Post(fmt.Sprintf("http://%v/%v", volumeAddr, fid), m.FormDataContentType(), read)
	if err != nil {
		log.Println("Upload file fails.", err)
	}
	return err
}

func VolumeDownloadFile(volumeAddr string, fid string, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		log.Println("Open file fails.", err)
		return err
	}
	defer file.Close()

	resp, err := http.Get(fmt.Sprintf("http://%v/%v", volumeAddr, fid))
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

type ClgSearchRequest struct {
	Fid              string   `json:"fid"`
	NumSegments      uint64   `json:"nseg"`
	ArchiveID        string   `json:"archid"`
	UncompressedSize uint64   `json:"uncompressed_size"`
	Size             uint64   `json:"size"`
	Args             []string `json:"args"`
}

func ClgSearch(volumeAddr string, request ClgSearchRequest) ([]string, error) {
	// Generate json request
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		log.Println("Generate json search request fails.", err)
		return nil, err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%v/clgsearch", volumeAddr), bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Println("Generate search request fails.", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Send clg search request fails.", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Read clg search result fails.", err)
	}
	return strings.Split(string(body), "\n"), nil
}

func ClgRemoteSearch(volumeAddr string, request ClgSearchRequest) ([]string, error) {
	// Generate json request
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		log.Println("Generate json search request fails.", err)
		return nil, err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%v:8000/clgremotesearch", volumeAddr), bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Println("Generate search request fails.", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Send clg search request fails.", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Read clg search result fails.", err)
	}
	return strings.Split(string(body), "\n"), nil
}
