package filesystem

import (
	"html/template"
	"net/http"

	"github.com/GeertJohan/go.rice"
)

//go:generate counterfeiter . FileSystemHelper

var (
	homepageTemplate *template.Template
)

type FileSystemHelper interface {
	GetStaticFileSystem() (http.FileSystem, error)
	GetHomepageTemplate() (*template.Template, error)
}

type fileSystemHelper struct {
	staticFileSystem http.FileSystem
	templatesBox     *rice.Box
}

func NewFileSystemHelper(
	assetsDir string,
) FileSystemHelper {
	return &fileSystemHelper{
		templatesBox:     rice.MustFindBox(assetsDir + "/templates"),
		staticFileSystem: rice.MustFindBox(assetsDir + "/static").HTTPBox(),
	}
}

func (h *fileSystemHelper) GetStaticFileSystem() (http.FileSystem, error) {
	return h.staticFileSystem, nil
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
	templateString, err := h.templatesBox.String("homepage.html.tmpl")
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
