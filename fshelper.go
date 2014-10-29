package garagepi

import (
	"bytes"
	"io"
	"net/http"

	"github.com/GeertJohan/go.rice"
)

type FsHelper interface {
	GetStaticFileSystem() (http.FileSystem, error)
	GetHomepageTemplateContents() ([]byte, error)
}

type FsHelperImpl struct {
	staticFileSystem    http.FileSystem
	templatesFileSystem http.FileSystem
}

func NewFsHelperImpl(
	assetsDir string,
) *FsHelperImpl {
	return &FsHelperImpl{
		templatesFileSystem: rice.MustFindBox(assetsDir + "/templates").HTTPBox(),
		staticFileSystem:    rice.MustFindBox(assetsDir + "/static").HTTPBox(),
	}
}

func (h *FsHelperImpl) GetStaticFileSystem() (http.FileSystem, error) {
	return h.staticFileSystem, nil
}

func (h *FsHelperImpl) GetHomepageTemplateContents() ([]byte, error) {
	fs, err := h.getTemplatesFileSystem()
	if err != nil {
		return nil, err
	}
	f, err := fs.Open("homepage.html")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, f)
	return buf.Bytes(), err
}

func (h *FsHelperImpl) getTemplatesFileSystem() (http.FileSystem, error) {
	return h.templatesFileSystem, nil
}
