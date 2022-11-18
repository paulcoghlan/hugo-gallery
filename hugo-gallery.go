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
	"strings"
	"syscall"
	"text/template"
	"time"
)

var postTemplate string = `---
title: {{.Title}}
date: "{{.Date}}"
type: "gallery"
cover: {{ .Cover }}
---
`

type GalleryItem struct {
	Title string
	Date  string
	Cover string
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
		fmt.Printf("Usage: %s <Source Path> <Destination Section> <Title> [BaseUrl]\n", os.Args[0])
		syscall.Exit(1)
	}

	sourcePath := os.Args[1] + "/"
	section := os.Args[2] + "/"
	title := os.Args[3]
	var baseUrl = ""
	if len(os.Args) > 4 {
		baseUrl = os.Args[4]
	}
	contentPath := "content/" + section

	src, err := os.Stat(contentPath)
	if err != nil || !src.IsDir() {
		err = os.Mkdir(contentPath, 0755)
		if err != nil {
			fmt.Printf("content directory not found! Are you in a hugo directory?\n")
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
		// Ignore directories and .DS files
		if !file.IsDir() && strings.Index(file.Name(), ".") > 0 {
			coverImage = file.Name()
			err = copyFile(sourcePath+file.Name(), contentPath+file.Name())
			if err != nil {
				fmt.Printf("Failed to copy %s\n", file.Name())
				os.Exit(1)
			}
			dateTimeOriginal := readPhotoDate(contentPath + file.Name())
			if dateTimeOriginal.After(latestModified) {
				latestModified = dateTimeOriginal
			}
		}
	}
	generateExif(contentPath)
	generateGallery(sourcePath, contentPath, title, coverImage, latestModified, section, baseUrl)
}

func generateExif(contentPath string) {
	outputPattern := fmt.Sprintf("-w %s%%f.%%e.json", contentPath)
	// fmt.Printf("outputPattern %s\n", outputPattern)
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("/usr/bin/exiftool %s -json -struct -EXIF:All -XMP:Title -Composite:LensSpec %s", outputPattern, contentPath))

	fmt.Printf("running %s\n", cmd)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func generateGallery(sourcePath string, contentPath string, title string, coverImage string, date time.Time, section string, baseUrl string) {
	galleryItem := GalleryItem{
		Title: title,
		Date:  date.Format("2006-01-02"),
		Cover: coverImage,
	}

	var buffer bytes.Buffer
	generateTemplate(galleryItem, &buffer)

	filePath := contentPath + "index.md"
	fmt.Printf("Create markdown %s\n", filePath)
	f, err := os.Create(filePath)
	check(err)
	defer f.Close()
	f.Sync()
	w := bufio.NewWriter(f)
	w.WriteString(buffer.String())
	w.Flush()
}

func generateTemplate(galleryItem GalleryItem, buffer *bytes.Buffer) {
	t := template.New("post template")
	t, err := t.Parse(postTemplate)
	check(err)
	err = t.Execute(buffer, galleryItem)
	check(err)
}

func readPhotoDate(sourcePath string) time.Time {
	layout := "2006-01-02 15:04:05"
	tag := "DateTimeOriginal"
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("/usr/bin/exiftool -S -%s -d '%%Y-%%m-%%d %%H:%%M:%%S' %s", tag, sourcePath))
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
	fmt.Printf("Copy file %s to %s\n", in, out)
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
