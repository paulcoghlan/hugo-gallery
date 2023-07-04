#!/usr/bin/env bash

echo "Importing /mnt/d/to-import/$1"

HUGO_DIR=/mnt/d/pixse1
ls /mnt/d/to-import/$1 | xargs -I {} ./hugo-gallery /mnt/d/to-import/$1/{} gallery/smugmug/$1/{} {}
