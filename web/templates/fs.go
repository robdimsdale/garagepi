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

	"/templates/head.html.tmpl": {
		local: "web/assets/templates/head.html.tmpl",
		size:  726,
		compressed: `
H4sIAAAJbogA/5RSwY7TMBC95yuMzySm5cIhjoRKhTjBASQ4uvY0cXBs156mVFH/HTfe7nZ3tautcsjM
y7z3lDczTQq22gKhHQhFT6eifvfl++rnnx9r0uFgmqI+v4gRtuUULG0KQurz7LlI5QAoiOxEiICc7nFb
fqLXnzpEX8Jur0dOf5e/PpcrN3iBemOAEuksgk28b2sOqoVHTCsG4HTUcPAu4NXwQSvsuIJRSyjn5j3R
VqMWpoxSGOCLixBqNNB8FUG0QFZJIDhTs4wWeSTKoD0SPPrkhvAPWS9GkVFKYpCcMiadgqrf7SEcK+kG
lstyWS3SM2hb9ZE2Ncus5gZhC6isqDbOYcQgvFR2NrgH2MdqWX1gfXyAXjI02v4lAQynEY8GYgeQjLoA
21ucZHxqlRB6SetVj4hpsXJWEN4bLVPrMv86hqz35ozuVFMC7bxHr5/9fM3yRU4TWJVu+H8AAAD//5g8
zzzWAgAA
`,
	},

	"/templates/homepage.html.tmpl": {
		local: "web/assets/templates/homepage.html.tmpl",
		size:  1311,
		compressed: `
H4sIAAAJbogA/7RUzY7TMBC+9ykGn+BgQpcVQkuaC0gIUakHeAEncRIL2xPZE7pVlHdn4nTLlgpEQXuo
Op6fb+b7PPE41roxXoPo0OletVpM02ocSbveKpr9WtWzDyAvsT4Aeouq3ogqaI5/clyyVQcdnr8QBSdx
Wm2+Q2VVjJyEnhTDh2PsPBpwf/L/WmflfZTrm0dxzujWxUcVuCO8Z+CANs/Y9RMhY4hTo+Xwd20NE9rr
slJOXI4AkQ5Wc4Kpqbt7ffOqv3/XadN2dLd+y4ezGRnQuBZiqDYie4BcsjeC0wUkmI1gHAHZH4eH/JmU
wOOClP+oH8xWdPJNMlwtb5NhW3l7Lm05EKFPQpTkPyCGr9i2Vp/0YC/wT/K+qMFSskuL1bdkqYoMelEs
RTDX59mCeTXFcQTTwMsvxPv12ePeQ1q/J2R/wX8739e11IfgefZ59FS+89O0axpmo23UzGHnx1H7epog
xf9HH4Z5clHyBoMDp6lDFqXHyIIsXHmzLbY40G9EpEPPn0scSmeuFDHp/4C9Tf+XMs3azLNdKdxj/+ld
Oka5Cb9txYrfE3K2WB0vavUjAAD//7LeXwUfBQAA
`,
	},

	"/templates/login.html.tmpl": {
		local: "web/assets/templates/login.html.tmpl",
		size:  853,
		compressed: `
H4sIAAAJbogA/6RTS27jMAzdzykE7Q0jM8GsYp9gFrPpAWSLjoVKoiHRTQIjdy9lO/6kKFqgCxtP4iP5
nkQNg4bGeBDS4tl4eb//GgYC11lFvNmC0mlPiFOF+lYyYKjNm6itirGQNXpSnB/kFNtHA16W/ec8m11j
dvgtEoou+/sA2DQRKPszrp3Ojg8wB46bilyzPZT/kvJTzmgbaDA44YBa1IXsMJIUqiaDvpD55HVL36tL
udk5YN89kZhmVQVWMKOQXjmQ5UuEkNApH0MfEozvehJ066CQBFfWse2SDjCglcLouaBI/wfme6ihRath
6bdXnbPsHxvpmHzBoGX5f0bfMbNkfW5opUym1vXO2Nr/S3NVT4R+VhD7ypn1QCvygr+sC8apcBtxZbF+
ncTMtz6Py1RoNzJ5MrCZ1237zWKBXGR8FDx75PishgG85tfyHgAA//9yVpMzVQMAAA==
`,
	},

	"/": {
		isDir: true,
		local: "web/assets",
	},

	"/templates": {
		isDir: true,
		local: "web/assets/templates",
	},
}
