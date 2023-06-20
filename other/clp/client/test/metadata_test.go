package test

import (
	"log"
	"reflect"
	"sort"
	"testing"
)
import "clp_client/metadata"

var archives = []metadata.ArchiveMetadata{
	{UncompressedSize: 100, Size: 10, Fid: "3,130fa903"},
	{UncompressedSize: 250, Size: 30, Fid: "1,369b901ac7"},
}

var files = []metadata.FileMetadata{
	{"/path/to/0", "tag1", 4000, 5000, 10000, 500, 0},
	{"/path/to/1", "tag2", 10000, 12000, 5000, 100, 1},
	{"/path/to/2", "tag1", 7000, 8000, 8000, 400, 0},
	{"/path/to/3", "tag2", 9000, 11000, 3000, 100, 0},
}

var db metadata.MetadataService

func init() {
	db = &metadata.MetaMySQL{}
	err := db.Connect("test", "pwd", "127.0.0.1:3306")
	if err != nil {
		log.Fatal("Connect to metadata service fail.", err)
	}
	db.InitService()
	db.AddMetadata(archives, files)
}

func ListEqual(l1 []string, l2 []string) bool {
	sort.Strings(l1)
	sort.Strings(l2)
	return reflect.DeepEqual(l1, l2)
}

func TestGetMeta(t *testing.T) {
	tags, err := db.ListTags()
	if err != nil {
		t.Fatal("Get tag fail.", err)
	}
	tagsExpected := []string{"tag1", "tag2"}
	if !ListEqual(tags, tagsExpected) {
		t.Fatalf("Get tag not match. Expect %v. Get %v.\n", tagsExpected, tags)
	}
	files, err := db.GetFiles("tag1")
	if err != nil {
		t.Fatal("Get file fail.", err)
	}
	filesExpected := []string{"/path/to/0", "/path/to/2"}
	if !ListEqual(files, filesExpected) {
		t.Fatalf("Get file not match. Expect %v. Get %v.\n", filesExpected, files)
	}
}

func TestSearch(t *testing.T) {
	result_1, err := db.Search("tag1", 4500, 6000)
	if err != nil {
		t.Fatal("Search fail.", err)
	}
	resultExpected_1 := []string{"3,130fa903"}
	if !ListEqual(result_1, resultExpected_1) {
		t.Fatalf("Search not match. Expect %v. Get %v.\n", result_1, resultExpected_1)
	}
	result_2, err := db.Search("tag2", 9000, 11000)
	if err != nil {
		t.Fatal("Search fail.", err)
	}
	resultExpected_2 := []string{"3,130fa903", "1,369b901ac7"}
	if !ListEqual(result_2, resultExpected_2) {
		t.Fatalf("Search not match. Expect %v. Get %v.\n", result_2, resultExpected_2)
	}
}
