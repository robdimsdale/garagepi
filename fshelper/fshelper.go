package fshelper

import (
	"html/template"
	"net/http"

	"github.com/GeertJohan/go.rice"
)

//go:generate counterfeiter . FsHelper

var (
	homepageTemplate *template.Template
)

type FsHelper interface {
	GetStaticFileSystem() (http.FileSystem, error)
	GetHomepageTemplate() (*template.Template, error)
}

type fsHelperImpl struct {
	staticFileSystem http.FileSystem
	templatesBox     *rice.Box
}

func NewFsHelperImpl(
	assetsDir string,
) FsHelper {
	return &fsHelperImpl{
		templatesBox:     rice.MustFindBox(assetsDir + "/templates"),
		staticFileSystem: rice.MustFindBox(assetsDir + "/static").HTTPBox(),
	}
}

func (h *fsHelperImpl) GetStaticFileSystem() (http.FileSystem, error) {
	return h.staticFileSystem, nil
}

func (h *fsHelperImpl) GetHomepageTemplate() (*template.Template, error) {
	if homepageTemplate == nil {
		err := h.loadHomepageTemplate()
		if err != nil {
			return nil, err
		}
	}
	return homepageTemplate, nil
}

func (h *fsHelperImpl) loadHomepageTemplate() error {
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
