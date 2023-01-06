# [Hugo](http://hugo.spf13.com) image gallery generator

This tool will create a new posts directory containing a markdown file for each image in source directory allowing for an ordered slide show.

## Usage
`hugo-gallery <Source Path> <Destination Section> <Title> [BaseUrl]`

## Example

`hugo-gallery static/images/vacation-photos hawaii "Hawaii Trip"`

Visit `localhost:1313/hawaii` to view the content.

This would read all of the images out of the `static/images/vacation-photos` directory and create a new folder named `hawaii` in `content/hawaii` filled with front matter markdown files. See sample below for details.

### Markdown Sample

```yml
---
---
```

## Todo:

## License
* MIT
