package weed

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
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

func DownloadFile(volumeAddr string, fid string, filepath string) error {
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
