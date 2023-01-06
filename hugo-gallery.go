package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"
)

const slash = string(os.PathSeparator)

var galleryTemplate string = `---
title: {{.Title}}
date: "{{.Date}}"
type: "gallery"
cover: "{{ .Cover }}"
---
`

var collectionTemplate string = `---
title: {{.Title}}
date: "{{.Date}}"
type: "collection"
cover: "{{ .Cover }}"
---
`

// e.g. "images/colin-watts-5yHnRGgk6is-unsplash.jpg"
type PostItem struct {
	Title          string
	TitleLowerCase string
	Date           string
	Cover          string
}

func check(e error) int {
	result := 0
	if e != nil {
		result = 1
		defer func() {
			panic(e)
		}()
	}
	return result
}

func main() {
	if len(os.Args) < 4 {
		// e.g. hugo-gallery /mnt/d/photos/2022/hawaii gallery/2022/hawaii "Hawaii Trip"
		fmt.Printf("Usage: HUGO_DIR=$HOME/personal/pixse1 %s <Source Path> <Destination Section> <Title>\n", os.Args[0])
		syscall.Exit(1)
	}

	hugoDir := os.Getenv("HUGO_DIR")
	assetsDir := filepath.Join(hugoDir, "assets")
	sourcePath := os.Args[1] + slash
	section := os.Args[2] + slash
	title := os.Args[3]
	contentPath := filepath.Join(hugoDir, "content", section)

	src, err := os.Stat(contentPath)
	if err != nil || !src.IsDir() {
		err = os.Mkdir(contentPath, 0755)
		if err != nil {
			fmt.Printf("content directory <%s> not found! Are you in a hugo directory?\n", contentPath)
			os.Exit(1)
		}
	}

	imgList, err := ioutil.ReadDir(sourcePath)
	if err != nil {
		fmt.Printf("Source path <%s> not found!\n", sourcePath)
		os.Exit(1)
	}

	coverImage := ""
	latestModified := time.Date(1970, 0, 0, 0, 0, 0, 0, time.UTC)
	for _, file := range imgList {
		fileName := strings.ToLower(file.Name())
		// Ignore directories and .DS files
		if !file.IsDir() && strings.Index(fileName, ".") > 0 {
			if strings.Index(fileName, ".jpg") == -1 {
				continue
			}
			coverImage = file.Name()
			err = copyFile(filepath.Join(sourcePath, file.Name()), filepath.Join(contentPath, file.Name()))
			if err != nil {
				fmt.Printf("Failed to copy %s\n", file.Name())
				os.Exit(1)
			}
			dateTimeOriginal := readPhotoDate(filepath.Join(contentPath, file.Name()))
			if dateTimeOriginal.After(latestModified) {
				latestModified = dateTimeOriginal
			}
		}
	}
	generateGallery(contentPath, title, coverImage, latestModified)

	collectionExists := parentPost(contentPath)
	fmt.Printf("collectionExists is %v\n", collectionExists)
	if !collectionExists {
		collectionDir := parentDir(contentPath)
		baseDir := filepath.Dir(collectionDir)
		collectionName := filepath.Base(collectionDir)
		generateCollection(sourcePath, assetsDir, baseDir, title, coverImage, latestModified, collectionName)
	}
}

func parentDir(contentPath string) string {
	parent, _ := filepath.Split(contentPath)
	parentDir := filepath.Dir(parent)
	fmt.Printf("parentDir %s\n", parentDir)
	return parentDir
}

func parentPost(contentPath string) bool {
	parentDir := parentDir(contentPath)
	parentMD := parentDir + ".md"
	md, err := os.Stat(parentMD)
	fmt.Printf("parentMD is %s, md is: %v\n", parentMD, md)
	if err != nil {
		return false
	}
	return true
}

func generateGallery(contentPath string, title string, coverImage string, date time.Time) {
	galleryItem := PostItem{
		Title: title,
		Date:  date.Format("2006-01-02"),
		Cover: coverImage,
	}

	var buffer bytes.Buffer
	generateGalleryPost(galleryItem, &buffer)

	filePath := filepath.Join(contentPath, "index.md")
	fmt.Printf("Create markdown %s\n", filePath)
	f, err := os.Create(filePath)
	check(err)
	defer f.Close()
	f.Sync()
	w := bufio.NewWriter(f)
	w.WriteString(buffer.String())
	w.Flush()
}

func generateGalleryPost(galleryItem PostItem, buffer *bytes.Buffer) {
	t := template.New("gallery template")
	t, err := t.Parse(galleryTemplate)
	check(err)
	err = t.Execute(buffer, galleryItem)
	check(err)
}

func generateCollection(sourcePath string, assetsDir string, parentPath string, title string, coverImage string, date time.Time, collectionName string) {
	contentImage := fmt.Sprintf("images/%s-%s", strings.ToLower(collectionName), coverImage)
	collectionItem := PostItem{
		Title: collectionName,
		Date:  date.Format("2006-01-02"),
		Cover: contentImage,
	}

	// Copy coverImage to /images/collectionName-imagename
	err := copyFile(filepath.Join(sourcePath, coverImage), filepath.Join(assetsDir, contentImage))
	check(err)

	var buffer bytes.Buffer
	generateCollectionPost(collectionItem, &buffer)

	filePath := filepath.Join(parentPath, collectionName+".md")
	fmt.Printf("Create markdown %s\n", filePath)
	f, err := os.Create(filePath)
	check(err)
	defer f.Close()
	f.Sync()
	w := bufio.NewWriter(f)
	w.WriteString(buffer.String())
	w.Flush()
}

func generateCollectionPost(galleryItem PostItem, buffer *bytes.Buffer) {
	t := template.New("collection template")
	t, err := t.Parse(collectionTemplate)
	check(err)
	err = t.Execute(buffer, galleryItem)
	check(err)
}

func readPhotoDate(sourcePath string) time.Time {
	layout := "2006-01-02 15:04:05"
	tag := "DateTimeOriginal"
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("/opt/homebrew/bin/exiftool -S -%s -d '%%Y-%%m-%%d %%H:%%M:%%S' %s", tag, sourcePath))
	fmt.Printf("running %s\n", cmd)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	split := strings.SplitAfter(string(output), tag+":")
	dateString := strings.TrimSpace(split[1])
	t, err := time.Parse(layout, dateString)
	if err != nil {
		fmt.Println("Error while parsing date :", err)
	}

	return t
}

func copyFile(in string, out string) error {
	fmt.Printf("Copy file: %s to: %s\n", in, out)
	srcFile, err := os.Open(in)
	check(err)
	defer srcFile.Close()

	destFile, err := os.Create(out)
	check(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	check(err)

	err = destFile.Sync()
	check(err)

	return nil
}
