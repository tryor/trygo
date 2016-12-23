//Initial code source, @see https://github.com/astaxie/beego/blob/master/template.go

package trygo

import (
	"errors"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type TemplateRegister struct {
	app         *App
	tplFuncMap  template.FuncMap
	Templates   map[string]*template.Template
	TemplateExt []string
}

func NewTemplateRegister(app *App) *TemplateRegister {
	tr := &TemplateRegister{app: app}
	tr.Templates = make(map[string]*template.Template)
	tr.tplFuncMap = make(template.FuncMap)
	tr.TemplateExt = make([]string, 0)
	tr.TemplateExt = append(tr.TemplateExt, "tpl", "html")

	tr.tplFuncMap["dateformat"] = dateFormat
	tr.tplFuncMap["date"] = date
	tr.tplFuncMap["compare"] = compare
	tr.tplFuncMap["compare_not"] = compareNot
	tr.tplFuncMap["not_nil"] = notNil
	tr.tplFuncMap["not_null"] = notNil
	tr.tplFuncMap["substr"] = substr
	tr.tplFuncMap["html2str"] = html2str
	tr.tplFuncMap["str2html"] = str2html
	tr.tplFuncMap["htmlquote"] = htmlquote
	tr.tplFuncMap["htmlunquote"] = htmlunquote
	tr.tplFuncMap["renderform"] = renderForm
	tr.tplFuncMap["assets_js"] = assetsJs
	tr.tplFuncMap["assets_css"] = assetsCSS
	//tr.tplFuncMap["config"] = GetConfig
	tr.tplFuncMap["map_get"] = mapGet

	// Comparisons
	tr.tplFuncMap["eq"] = eq // ==
	tr.tplFuncMap["ge"] = ge // >=
	tr.tplFuncMap["gt"] = gt // >
	tr.tplFuncMap["le"] = le // <=
	tr.tplFuncMap["lt"] = lt // <
	tr.tplFuncMap["ne"] = ne // !=

	return tr
}

// AddFuncMap let user to register a func in the template
func (this *TemplateRegister) AddFuncMap(key string, funname interface{}) error {
	if _, ok := this.tplFuncMap[key]; ok {
		return errors.New("funcmap already has the key")
	}
	this.tplFuncMap[key] = funname
	return nil
}

type templatefile struct {
	root  string
	files map[string][]string
}

func (self *templatefile) visit(tr *TemplateRegister, paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() {
		return nil
	} else if (f.Mode() & os.ModeSymlink) > 0 {
		return nil
	} else {
		hasExt := false
		for _, v := range tr.TemplateExt {
			if strings.HasSuffix(paths, v) {
				hasExt = true
				break
			}
		}
		if hasExt {
			replace := strings.NewReplacer("\\", "/")
			a := []byte(paths)
			a = a[len([]byte(self.root)):]
			subdir := path.Dir(strings.TrimLeft(replace.Replace(string(a)), "/"))
			if _, ok := self.files[subdir]; ok {
				self.files[subdir] = append(self.files[subdir], paths)
			} else {
				m := make([]string, 1)
				m[0] = paths
				self.files[subdir] = m
			}

		}
	}
	return nil
}

func (this *TemplateRegister) AddTemplateExt(ext string) {
	for _, v := range this.TemplateExt {
		if v == ext {
			return
		}
	}
	this.TemplateExt = append(this.TemplateExt, ext)
}

func (this *TemplateRegister) buildTemplate(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return err
		} else {
			return errors.New("dir open error")
		}
	}
	self := templatefile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return self.visit(this, path, f, err)
	})
	if err != nil {
		this.app.Logger.Error("filepath.Walk() returned %v", err)
		return err
	}
	for k, v := range self.files {
		this.Templates[k] = template.Must(template.New("template" + k).Funcs(this.tplFuncMap).ParseFiles(v...))
	}
	return nil
}
