//Initial code source, @see https://github.com/astaxie/beego/blob/master/template.go

package trygo

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	//	"path"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func executeTemplate(app *App, wr io.Writer, name string, data interface{}) error {
	if app.Config.RunMode == DEV {
		app.TemplateRegister.templatesLock.RLock()
		defer app.TemplateRegister.templatesLock.RUnlock()
	}
	if t, ok := app.TemplateRegister.Templates[name]; ok {
		var err error
		if t.Lookup(name) != nil {
			err = t.ExecuteTemplate(wr, name, data)
		} else {
			err = t.Execute(wr, data)
		}
		if err != nil {
			log.Println("template Execute err:", err)
		}
		return err
	}
	panic("can't find templatefile in the path:" + name)
}

type TemplateRegister struct {
	app           *App
	tplFuncMap    template.FuncMap
	Templates     map[string]*template.Template
	TemplateExt   []string
	templatesLock sync.RWMutex

	templateEngines map[string]templatePreProcessor
}

func NewTemplateRegister(app *App) *TemplateRegister {
	tr := &TemplateRegister{app: app}
	tr.Templates = make(map[string]*template.Template)
	tr.tplFuncMap = make(template.FuncMap)
	tr.TemplateExt = make([]string, 0)
	tr.TemplateExt = append(tr.TemplateExt, "tpl", "html")
	tr.templateEngines = map[string]templatePreProcessor{}
	//beegoTplFuncMap = make(template.FuncMap)

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

type templatePreProcessor func(root, path string, funcs template.FuncMap) (*template.Template, error)

type templatefile struct {
	root  string
	files map[string][]string
}

func (tf *templatefile) visit(tr *TemplateRegister, paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	if !hasTemplateExt(tr, paths) {
		return nil
	}

	replace := strings.NewReplacer("\\", "/")
	file := strings.TrimLeft(replace.Replace(paths[len(tf.root):]), "/")
	subDir := filepath.Dir(file)

	tf.files[subDir] = append(tf.files[subDir], file)
	return nil
}

func hasTemplateExt(tr *TemplateRegister, paths string) bool {
	for _, v := range tr.TemplateExt {
		if strings.HasSuffix(paths, "."+v) {
			return true
		}
	}
	return false
}

//func (self *templatefile) visit(tr *TemplateRegister, paths string, f os.FileInfo, err error) error {
//	if f == nil {
//		return err
//	}
//	if f.IsDir() {
//		return nil
//	} else if (f.Mode() & os.ModeSymlink) > 0 {
//		return nil
//	} else {
//		hasExt := false
//		for _, v := range tr.TemplateExt {
//			if strings.HasSuffix(paths, v) {
//				hasExt = true
//				break
//			}
//		}
//		if hasExt {
//			replace := strings.NewReplacer("\\", "/")
//			a := []byte(paths)
//			a = a[len([]byte(self.root)):]
//			fmt.Println("a:", string(a))
//			subdir := path.Dir(strings.TrimLeft(replace.Replace(string(a)), "/"))
//			fmt.Println("subdir:", (subdir))
//			if _, ok := self.files[subdir]; ok {
//				self.files[subdir] = append(self.files[subdir], paths)
//			} else {
//				m := make([]string, 1)
//				m[0] = paths
//				self.files[subdir] = m
//			}

//		}
//	}
//	return nil
//}

func (this *TemplateRegister) AddTemplateExt(ext string) {
	for _, v := range this.TemplateExt {
		if v == ext {
			return
		}
	}
	this.TemplateExt = append(this.TemplateExt, ext)
}

func (this *TemplateRegister) buildTemplate(dir string, files ...string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New("dir open err")
	}
	self := &templatefile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return self.visit(this, path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}

	buildAllFiles := len(files) == 0
	for _, v := range self.files {
		for _, file := range v {
			if buildAllFiles || inSlice(file, files) {
				func() {
					this.templatesLock.Lock()
					defer this.templatesLock.Unlock()
					ext := filepath.Ext(file)
					var t *template.Template
					if len(ext) == 0 {
						t, err = this.getTemplate(self.root, file, v...)
					} else if fn, ok := this.templateEngines[ext[1:]]; ok {
						t, err = fn(self.root, file, this.tplFuncMap)
					} else {
						t, err = this.getTemplate(self.root, file, v...)
					}
					if err != nil {
						//logs.Trace("parse template err:", file, err)
						log.Println("parse template err:", file, err)
					} else {
						this.Templates[file] = t
					}
					//this.templatesLock.Unlock()
				}()
			}
		}
	}
	return nil
}

//func (this *TemplateRegister) buildTemplate(dir string) error {
//	if _, err := os.Stat(dir); err != nil {
//		if os.IsNotExist(err) {
//			return err
//		} else {
//			return errors.New("dir open error")
//		}
//	}
//	self := templatefile{
//		root:  dir,
//		files: make(map[string][]string),
//	}
//	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
//		fmt.Println(dir, path)
//		return self.visit(this, path, f, err)
//	})
//	if err != nil {
//		this.app.Logger.Error("filepath.Walk() returned %v", err)
//		return err
//	}
//	for k, v := range self.files {
//		this.Templates[k] = template.Must(template.New("template" + k).Funcs(this.tplFuncMap).ParseFiles(v...))
//	}
//	return nil
//}

func (this *TemplateRegister) getTemplate(root, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file).Delims(this.app.Config.TemplateLeft, this.app.Config.TemplateRight).Funcs(this.tplFuncMap)
	var subMods [][]string
	t, subMods, err = this.getTplDeep(root, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = this._getTemplate(t, root, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func (this *TemplateRegister) _getTemplate(t0 *template.Template, root string, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			//first check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = this.getTplDeep(root, otherFile, "", t)
					if err != nil {
						//logs.Trace("template parse file err:", err)
						log.Println("template parse file err:", err)
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = this._getTemplate(t, root, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				fileAbsPath := filepath.Join(root, otherFile)
				data, err := ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile(this.app.Config.TemplateLeft + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = this.getTplDeep(root, otherFile, "", t)
						if err != nil {
							log.Println("template parse file err:", err)
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = this._getTemplate(t, root, subMods1, others...)
						}
						break
					}
				}
			}
		}

	}
	return
}

func (this *TemplateRegister) getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	if filepath.HasPrefix(file, "../") {
		fileAbsPath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		fileAbsPath = filepath.Join(root, file)
	}
	if e := fileExists(fileAbsPath); !e {
		panic("can't find template file:" + file)
	}
	data, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile(this.app.Config.TemplateLeft + "[ ]*template[ ]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			tl := t.Lookup(m[1])
			if tl != nil {
				continue
			}
			if !hasTemplateExt(this, m[1]) {
				continue
			}
			t, _, err = this.getTplDeep(root, m[1], file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

//func fileExists(name string) bool {
//	if _, err := os.Stat(name); err != nil {
//		if os.IsNotExist(err) {
//			return false
//		}
//	}
//	return true
//}

func inSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}
