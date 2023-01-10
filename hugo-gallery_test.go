package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func Test_getTaken(t *testing.T) {
	type args struct {
		name string
	}
	taken := time.Unix(1653898754, 0)
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "jpg read",
			args: args{
				name: "test/source/a/b/c/d.jpg",
			},
			want: taken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTaken(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTaken() = %v, want %v", got.Unix(), tt.want)
			}
		})
	}
}

func shouldExist(t *testing.T, file string) {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		t.Errorf("file %s should exist but doesn't", file)
	}
}

func Test_importGallery(t *testing.T) {
	site := "./test/sample-site"
	err := os.RemoveAll(site)
	if err != nil {
		t.Errorf("RemoveAll() failed %v", err)
	}
	err = os.MkdirAll(filepath.Join(site, "content", "gallery"), 0755)
	if err != nil {
		t.Errorf("MkdirAll() failed %v", err)
	}
	err = os.MkdirAll(filepath.Join(site, "assets", "images"), 0755)
	if err != nil {
		t.Errorf("MkdirAll() failed %v", err)
	}
	section := "gallery/a/b/c"
	contentPath := filepath.Join(site, "content", section)
	importGallery(filepath.Join(site, "assets"), "./test/source/a/b/c", Gallery{
		title:       "test gallery",
		section:     section,
		contentPath: contentPath,
	})

	shouldExist(t, filepath.Join(contentPath, "d.jpg"))
	shouldExist(t, filepath.Join(site, "content", "gallery", "a.md"))
	shouldExist(t, filepath.Join(site, "content", "gallery", "a/b.md"))
	shouldExist(t, filepath.Join(site, "content", "gallery", "a/b/c/index.md"))
}
