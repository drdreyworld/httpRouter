package httpRouter

import (
	"html/template"
	"io"
)

var AppTemplates = Templates{}

type Templates struct {
	root  string
	paths []string
	tmpls *template.Template
}

func (t *Templates) SetRoot(root string) {
	t.root = root
}

func (t *Templates) AddPath(path string) {
	t.paths = append(t.paths, path)
}

func (t *Templates) ParseGlob() (err error) {
	if t.tmpls == nil {
		t.tmpls = template.New("templates")
		for i := 0; i < len(t.paths); i++ {
			_, err = t.tmpls.ParseGlob(t.root + t.paths[i])
			if err != nil {
				break
			}
		}
	}

	return
}

func (t *Templates) GetTemplates() *template.Template {
	return t.tmpls
}

func (t *Templates) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return t.GetTemplates().ExecuteTemplate(w, name, data)
}
