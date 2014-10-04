package oshelper

import (
	"net/http"
	"os/exec"

	"github.com/GeertJohan/go.rice"
)

type OsHelper interface {
	Exec(executable string, arg ...string) (string, error)
	GetStaticFileSystem() (http.FileSystem, error)
	GetHomepageTemplate() (http.File, error)
}

type OsHelperImpl struct {
	staticFileSystem    http.FileSystem
	templatesFileSystem http.FileSystem
}

func NewOsHelperImpl(
	assetsDir string,
) *OsHelperImpl {
	return &OsHelperImpl{
		templatesFileSystem: rice.MustFindBox(assetsDir + "/templates").HTTPBox(),
		staticFileSystem:    rice.MustFindBox(assetsDir + "/static").HTTPBox(),
	}
}

func (h *OsHelperImpl) Exec(executable string, arg ...string) (string, error) {
	out, err := exec.Command(executable, arg...).CombinedOutput()
	return string(out), err
}

func (h *OsHelperImpl) GetStaticFileSystem() (http.FileSystem, error) {
	return h.staticFileSystem, nil
}

func (h *OsHelperImpl) GetHomepageTemplate() (http.File, error) {
	fs, err := h.getTemplatesFileSystem()
	if err != nil {
		return nil, err
	}
	return fs.Open("homepage.html")
}

func (h *OsHelperImpl) getTemplatesFileSystem() (http.FileSystem, error) {
	return h.templatesFileSystem, nil
}
