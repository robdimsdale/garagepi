package oshelper

import (
	"bytes"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/robdimsdale/garage-pi/logger"
)

type OsHelper interface {
	Exec(executable string, arg ...string) (string, error)
	GetStaticFileSystem() (http.FileSystem, error)
	GetHomepageTemplateContents() ([]byte, error)
	Sleep(d time.Duration)
}

type OsHelperImpl struct {
	l                   logger.Logger
	staticFileSystem    http.FileSystem
	templatesFileSystem http.FileSystem
}

func NewOsHelperImpl(
	l logger.Logger,
	assetsDir string,
) *OsHelperImpl {
	return &OsHelperImpl{
		l:                   l,
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

func (h *OsHelperImpl) GetHomepageTemplateContents() ([]byte, error) {
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

func (h *OsHelperImpl) getTemplatesFileSystem() (http.FileSystem, error) {
	return h.templatesFileSystem, nil
}

func (h *OsHelperImpl) Sleep(d time.Duration) {
	h.l.Log("sleeping for " + d.String())
	time.Sleep(d)
}
