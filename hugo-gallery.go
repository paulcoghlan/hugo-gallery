package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

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

type Gallery struct {
	title       string
	section     string
	contentPath string
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
		fmt.Printf("Usage: HUGO_DIR=$HOME/mysite %s <Source Path> <Destination Section> <Title>\n", os.Args[0])
		syscall.Exit(1)
	}

	hugoDir := os.Getenv("HUGO_DIR")
	assetsDir := filepath.Join(hugoDir, "assets")
	sourcePath := filepath.Clean(os.Args[1])
	section := os.Args[2]
	title := os.Args[3]
	contentPath := filepath.Join(hugoDir, "content", section)
	importGallery(
		assetsDir,
		sourcePath,
		Gallery{
			title:       title,
			section:     section,
			contentPath: contentPath,
		})
}

func importGallery(assetsDir string, sourcePath string, gallery Gallery) {
	src, err := os.Stat(gallery.contentPath)
	if err != nil || !src.IsDir() {
		err = os.MkdirAll(gallery.contentPath, 0755)
		if err != nil {
			fmt.Printf("content directory <%s> not found! Are you in a hugo directory?\n", gallery.contentPath)
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
			if !strings.Contains(fileName, ".jpg") {
				continue
			}
			coverImage = file.Name()
			taken := getTaken(filepath.Join(sourcePath, file.Name()))
			if latestModified.Before(taken) {
				latestModified = taken
			}
			err = copyFile(filepath.Join(sourcePath, file.Name()), filepath.Join(gallery.contentPath, file.Name()))
			if err != nil {
				fmt.Printf("Failed to copy %s\n", file.Name())
				os.Exit(1)
			}
		}
	}
	generateGallery(gallery.contentPath, gallery.title, coverImage, latestModified)

	collectionExists := parentPost(gallery.contentPath)
	fmt.Printf("contentPath is %s, collectionExists is %v\n", gallery.contentPath, collectionExists)

	// Walk up directory tree from `contentPath` to ./gallery to see if we need to create a `/content/gallery/<collectionName>.md`
	collectionDir := parentDir(gallery.contentPath)
	for !collectionExists && (filepath.Base(collectionDir) != "gallery") {
		baseDir := filepath.Dir(collectionDir)
		collectionName := filepath.Base(collectionDir)
		fmt.Printf("collectionDir: %s, baseDir: %s, collectionName %s\n", collectionDir, baseDir, collectionName)
		generateCollection(sourcePath, assetsDir, baseDir, gallery.title, coverImage, latestModified, collectionName)
		collectionDir = parentDir(collectionDir)
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
	return err == nil
}

func generateGallery(contentPath string, title string, coverImage string, date time.Time) {
	galleryItem := PostItem{
		Title: cleanupTitle(title),
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
	err = f.Sync()
	check(err)
	w := bufio.NewWriter(f)
	_, err = w.WriteString(buffer.String())
	check(err)
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
		Title: cleanupTitle(collectionName),
		Date:  date.Format("2006-01-02"),
		Cover: contentImage,
	}
	fmt.Printf("collectionItem %v\n", collectionItem)

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
	err = f.Sync()
	check(err)
	w := bufio.NewWriter(f)
	_, err = w.WriteString(buffer.String())
	check(err)
	w.Flush()
}

func cleanupTitle(input string) string {
	return strings.Replace(input, "-", " ", -1)
}

func generateCollectionPost(galleryItem PostItem, buffer *bytes.Buffer) {
	t := template.New("collection template")
	t, err := t.Parse(collectionTemplate)
	check(err)
	err = t.Execute(buffer, galleryItem)
	check(err)
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

func getTaken(name string) time.Time {
	fmt.Printf("Get Exif on file: %s\n", name)
	f, err := os.Open(name)
	check(err)
	defer f.Close()

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(f)
	if err != nil && exif.IsCriticalError(err) {
		log.Fatal(err)
	}

	tm, _ := x.DateTime()
	fmt.Println("Taken: ", tm)
	return tm
}
