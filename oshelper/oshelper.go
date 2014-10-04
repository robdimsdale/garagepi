package oshelper

import "os/exec"

type OsHelper interface {
	Exec(executable string, arg ...string) (string, error)
}

type OsHelperImpl struct {
}

func NewOsHelperImpl() *OsHelperImpl {
	return &OsHelperImpl{}
}

func (h *OsHelperImpl) Exec(executable string, arg ...string) (string, error) {
	out, err := exec.Command(executable, arg...).CombinedOutput()
	return string(out), err
}
