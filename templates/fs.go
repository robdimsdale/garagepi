package templates

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDir struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDir) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Time{}
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDir{fs: _escLocal, name: name}
	}
	return _escDir{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/templates/homepage.html.tmpl": {
		local: "assets/templates/homepage.html.tmpl",
		size:  2917,
		compressed: `
H4sIAAAJbogA/7RWW2/bOhJ+96+Y6MlpIylOi92isVy0aREEWyRFmwLbR0ocSUwpUiUpO97A/32HpOxc
T097Dg6CwOTcL98MNd97f3Fy+e3TB2hdJxeTuf8ByVRTJKiSxQRg3iLj/kDHDh2DqmXGoiuSwdXpq+Qu
q3WuT/HHIJZF8t/069v0RHc9c6KUmECllUNFemcfCuQN3tNUrMMiWQpc9dq4O8IrwV1bcFyKCtNwOQCh
hBNMprZiEovZ1pATTuLilBnWIJyQAaPlPI/USRSxlRG9A7fuyZvDa5dfsSWL1ASsqYokzyvNMbv6MaBZ
Z5Xu8nhMj7IZ/XVCZVc2WczzqLX4DcMKHVcsK7V21hnWV1wFBztC/iI7yg7zK3tL+iOHUqjvYFAWiXVr
ibZFJEetwfp3PFX2oSuiJNtq/dSHddTYKlhgfS9FRVcd9e+WIdr75RqNVqkCTehjLx4lP8+3iJyXmq9B
K6kZL5LKIHN41pHWR7ZGM93fOuZiCZVk1pIQwYIJhWbk3ecavdrRH+rJ9Nqms6M7fD8bs0d4I9KthZxM
7BzFy6+5FZTQCsuKdcnjECC0Y5yN1y+ODvvr4xZF07rXs1d0uRcjGVR6rN5cdM1Y6Des8v0qrGK9bTU1
IKcq7ySfTuHPmriY5M+oEv3a+GBgerIPR4eH/4bPwi8NDm+dQ4PmAEwkkACLpIwQ651ctsJCb3RjWAd0
rA0iWF27FTN4DGs9QMUUgZILwqwoB4cgHDDFc22g01zUayJ4U4PiaMC1COShs6DrcDk9/wqnSAhgEj4N
JQEXPtJuURYPYInGUlHgKIMv5NeL10Kit3Zy8enb2fkp1OSG08YS0mbwLJ9MlsyA8KA7N1DA4THkOSkb
2k+ghq6kEMhxNRhD6ywKBpWalhgNFCcdhSt4awxbT/eD9mesqSSqQgtOg2+ZLq+wchZWLdUNWrbEW32u
V2EAhGqC4Z4NNpitmbR4PJnUgwqthscDAjeUWkygGQMJbAqEGETMAtKyXlsRTBSQsNJqSWVP7ov874zK
fU0C6WzLiINJpJD1RbjtlExFjJVQFH4m9bg8qPFOE87hOdASS+jnoUSrrfNvhZd4/ZSAfz6CepyeN6rw
UtPnz8cehcx8ypFPQXBdDR01J2vQfZDoj+/WZ3y6nT/SGIMO80YaNHHbPOLYEY0Gz9OiTiYITsa9Q0IL
TknuYMuohbHupBWSk9nNZELtvlxpenGpHxYI48Dkiq39DKD1iJnidYU0cMwFOBJC11BiI5Sihu8feISw
pRYcakLydzS37b5T9bHTjobrYbvGsgTghd7vkHUApSEf3kNtaLvt+fQocoTpDOY7sUyialwbPcTScpQe
f1sB24rajdh+jxJpZLXk0fPU7nvr3a4LwQgZyHrmJ+bcv8UGO73EULUpsUILN/S/89APtp367CJua5ju
xTHYfwL0vu7b3XZ/28VdB/O9NAXayZCmf/GRAH+yXfqvcOh4+jIcZJO+vP9+0P5yvlW07Uun3mttLnXT
hK+laJKoQP8px5oN0oVzSWD/Hk5xhyeLqARef55Hmz99hJ5K8ebGFy77Qg8w/kfRUoHN5p/N/lH+H/0s
/W7qg1EUuw89qF+ozeairikbpPVHOVyomxtUfLOBwP879SEz26Lcldt9U4zS5IK+S+hjOo8f1f8PAAD/
/7STH0FlCwAA
`,
	},

	"/": {
		isDir: true,
		local: "assets",
	},

	"/templates": {
		isDir: true,
		local: "assets/templates",
	},
}
