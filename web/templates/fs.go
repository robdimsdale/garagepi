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
		size:  2564,
		compressed: `
H4sIAAAJbogA/7RWXW/bNhe+96841ZXSxlKSt3g3NLaLNh2CYEFStN3FLinpSGJLkQJJ2fEM//c9pGy3
abJh2dALw+T5Ps/5oDabimupmZLWdNyLhpPtdrLZeO56JXygs6gCjWhWmGpNRisjqnlSWgb/qoPKtViz
TY+SBYQgVskllUo4ByGjvYB5u+Pd51qzOtC/11PTOzc9PfuGD4n2dHEpLDzSBQxbo2Y5SF8t5DBxcDRe
/plbiYRWXJSiSx6GQM6vFUNAVr599b+zk/7uvGXZtP7V6c+43IsRBrVxpZW9X8xk15Cz5TzJX4vSS6Pn
TovetcYnlC9m+UHy8RRwHfnk1z0C8Hzn889iKUYq/ObPgUS/tiEYSi+O6Ozk5Cf6IMtW2IreeM+W7THZ
kQABMZIyzT44+dRKR701jRUd4VhbZnKm9ith+ZzWZqBSaLJcSeetLAZ0hPQkdJUbS52pZL0GIZgadMWW
fMsED50jU8fL5c1vdMnoAKHo/VAoWdK1LFk7PqYlWwdQ6Cyjj/AbxGupOFi7uH3/+9XNJdVwUzF6SLmM
nueTyVJYkqHpbizN6eSc8hzKVsK8HroCIcBxOVjL2o+CUQU9Ll3LFXQ0r+iNtWKdHkXtD1wDEl2yI28o
lMwUn7n0jlYtcKNWLPmrfmVWcQCkbqLhXgwumq2Fcnw+mdSDjqWmhwNCG6Q2JtDsAolsBAIGiFnstKw3
TkYTc0pE4YwC7Ml9kT+uAPcdBKane8Y4mCDFrG/j7aBkSzBWUiP8TJlSBPMZCu8N+pxeUJLnCf6+l2iN
81p0HCRePSbQG+uj+jg9r/U8SKUvXuxqFDMLKY98BFGZcuhQnKxh/4vicHy7vqrS/fxBYxd0nDdoYOL2
eYxjBxoGL9BGnUyinax/y+gWTiF3vGfU0jp/0UpVwex2MkG5P60MqVAPR+hxEmol1mEG2IWOSfmuZAyc
8LEd0aFrKriRWqPgR8ehQ8TSyIpqdPIXtl/L/Q3qu0p7DNf35drBEhsv1v7QWcdUWPgIHmqL7fYspIfI
mdJTmh3EMsW68e3oYYS2YhX6by/gWln7XW+/Y8UYWaOq0XPqjoL17lCFaAQGsl6EibkxFWeWO7PkiFoK
Vixh2P8HD/3g2jRkN/ZtTemzcQyOHmn6gPt+t93fduOuo9mz6ZSwk2k6/ZePBIWT66b/j4eumr6MB9VM
X95/P7C/fCgVtn3h9Ttj7CfTNIoPSx9Uwm+KR1EMysdzgWb/Ek/jDk8WoxIF/Vk+2vzbR+ixFDebAFz2
0QOuXzWWCm23Pzb7B/lfh1l6auqD1Yg9hB7Vb/V2e1vXyIax/pDDrd5sWFfbLUX+f8EHZn44KDOsjI46
9q0BKFi8AGT3VCe5Mo0Z/F+AOD7Jbig6+UQQI/5729fx/yFMAZsQ2xOB+5Z++PjaceEEH3CLCT6afKcW
k12hJn8GAAD//yx6VJ8ECgAA
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
