package replay

import (
	"io"
	"regexp"
	"strings"

	"github.com/simonw/showboat/markdown"
)

var (
	htmlCommentRE = regexp.MustCompile(`<!--.*?-->`)
	showboatIDRE  = regexp.MustCompile(`<!--\s*showboat-id:.*?-->`)
	italicTimeRE  = regexp.MustCompile(`^\*[^*]+\*$`)
)

// ParseDocument reads a showboat markdown file and returns an ordered scene list.
func ParseDocument(r io.Reader) ([]Scene, error) {
	blocks, err := markdown.Parse(r)
	if err != nil {
		return nil, err
	}
	return blocksToScenes(blocks), nil
}

func blocksToScenes(blocks []markdown.Block) []Scene {
	var scenes []Scene

	for i := 0; i < len(blocks); i++ {
		switch b := blocks[i].(type) {
		case markdown.TitleBlock:
			if strings.TrimSpace(b.Title) != "" {
				scenes = append(scenes, TitleScene{Text: b.Title})
			}
		case markdown.CommentaryBlock:
			lines, hasHeading := commentaryLines(b.Text)
			if len(lines) > 0 {
				scenes = append(scenes, NoteScene{Lines: lines, HasHeading: hasHeading})
			}
		case markdown.CodeBlock:
			if b.IsImage {
				alt, path := b.Code, b.Code
				if i+1 < len(blocks) {
					if img, ok := blocks[i+1].(markdown.ImageOutputBlock); ok {
						alt = img.AltText
						path = img.Filename
						i++
					}
				}
				scenes = append(scenes, ImageScene{Alt: alt, Path: path})
			} else {
				output := ""
				if i+1 < len(blocks) {
					if ob, ok := blocks[i+1].(markdown.OutputBlock); ok {
						output = ob.Content
						i++
					}
				}
				scenes = append(scenes, ExecScene{
					Lang:    b.Lang,
					Command: b.Code,
					Output:  output,
				})
			}
		}
	}

	return scenes
}

func commentaryLines(text string) ([]string, bool) {
	text = htmlCommentRE.ReplaceAllString(text, "")
	text = showboatIDRE.ReplaceAllString(text, "")

	var lines []string
	hasHeading := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if italicTimeRE.MatchString(line) {
			continue
		}
		if strings.HasPrefix(line, "#") {
			hasHeading = true
		}
		lines = append(lines, line)
	}
	return lines, hasHeading
}
