package filesystem

import (
	"html/template"

	"github.com/robdimsdale/garagepi/templates"
)

//go:generate counterfeiter . FileSystemHelper

var (
	homepageTemplate *template.Template
)

type FileSystemHelper interface {
	GetHomepageTemplate() (*template.Template, error)
}

type fileSystemHelper struct {
}

func NewFileSystemHelper() FileSystemHelper {
	return &fileSystemHelper{}
}

func (h *fileSystemHelper) GetHomepageTemplate() (*template.Template, error) {
	if homepageTemplate == nil {
		err := h.loadHomepageTemplate()
		if err != nil {
			return nil, err
		}
	}
	return homepageTemplate, nil
}

func (h *fileSystemHelper) loadHomepageTemplate() error {
	templateString, err := templates.FSString(false, "/templates/homepage.html.tmpl")
	if err != nil {
		return err
	}

	// parse and execute the template
	tmplMessage, err := template.New("message").Parse(templateString)
	if err != nil {
		return err
	}

	homepageTemplate = tmplMessage

	return nil
}
