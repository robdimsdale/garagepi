package filesystem

import (
	"html/template"
	"path/filepath"

	"github.com/robdimsdale/garagepi/web/templates"
)

var (
	allTemplates *template.Template

	filenames = []string{
		"/templates/head.html.tmpl",
		"/templates/homepage.html.tmpl",
		"/templates/login.html.tmpl",
	}
)

func LoadTemplates() (*template.Template, error) {
	// Below is taken from http://golang.org/src/html/template/template.go
	// because there is no way to get all the templates from the in-memory filesystem.
	// We would like to use e.g. ParseFiles but that is hard-coded
	// to use an actual filesystem; we cannot retarget it.
	var t *template.Template
	for _, filename := range filenames {
		s, err := templates.FSString(false, filename)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(filename)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
