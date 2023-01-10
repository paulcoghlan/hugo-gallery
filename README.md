# [Hugo](http://hugo.spf13.com) gallery importer

This tool will copy over the image files from a directory into a given Hugo target directory and create a new markdown gallery post.

## Usage

`HUGO_DIR=<Hugo Site> hugo-gallery <Source Images Path> <Target Section> <Title>`

## Example

e.g.: `HUGO_DIR=$HOME/mysite hugo-gallery /Volumes/photos/photos/exports/2022/nice gallery/2022/nice "Nice Vacation"`

Visit `localhost:1313/gallery/2022/nice` to view the content.
