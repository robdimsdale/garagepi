package garagepi

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/GeertJohan/go.rice"
)

var (
	homepageTemplate *template.Template
)

type FsHelper interface {
	GetStaticFileSystem() (http.FileSystem, error)
	GetHomepageTemplateContents() ([]byte, error)
}

type FsHelperImpl struct {
	staticFileSystem http.FileSystem
	templatesBox     *rice.Box
}

func NewFsHelperImpl(
	assetsDir string,
) *FsHelperImpl {
	return &FsHelperImpl{
		templatesBox:     rice.MustFindBox(assetsDir + "/templates"),
		staticFileSystem: rice.MustFindBox(assetsDir + "/static").HTTPBox(),
	}
}

func (h *FsHelperImpl) GetStaticFileSystem() (http.FileSystem, error) {
	return h.staticFileSystem, nil
}

func (h *FsHelperImpl) GetHomepageTemplateContents() ([]byte, error) {
	if homepageTemplate == nil {
		err := h.loadHomepageTemplate()
		if err != nil {
			return nil, err
		}
	}

	buf := bytes.NewBuffer(nil)
	homepageTemplate.Execute(buf, map[string]string{"Message": "Hello, world!"})

	return buf.Bytes(), nil
}

func (h *FsHelperImpl) loadHomepageTemplate() error {
	templateString, err := h.templatesBox.String("homepage.html")
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
