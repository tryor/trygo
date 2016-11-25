//Initial code source, @see https://github.com/astaxie/beego/blob/master/template.go

package ssss

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TemplateRegister struct {
	tplFuncMap  template.FuncMap
	Templates   map[string]*template.Template
	TemplateExt []string
}

func NewTemplateRegister() *TemplateRegister {
	tr := &TemplateRegister{}
	tr.Templates = make(map[string]*template.Template)
	tr.tplFuncMap = make(template.FuncMap)
	tr.TemplateExt = make([]string, 0)
	tr.TemplateExt = append(tr.TemplateExt, "tpl", "html")
	//tr.tplFuncMap["markdown"] = MarkDown
	tr.tplFuncMap["dateformat"] = dateFormat
	tr.tplFuncMap["date"] = date
	tr.tplFuncMap["compare"] = compare
	tr.tplFuncMap["substr"] = substr
	tr.tplFuncMap["html2str"] = html2str
	return tr
}

//func MarkDown(raw string) (output template.HTML) {
//	input := []byte(raw)
//	bOutput := blackfriday.MarkdownBasic(input)
//	output = template.HTML(string(bOutput))
//	return
//}

func substr(s string, start, length int) string {
	bt := []rune(s)
	if start < 0 {
		start = 0
	}
	var end int
	if (start + length) > (len(bt) - 1) {
		end = len(bt) - 1
	} else {
		end = start + length
	}
	return string(bt[start:end])
}

func html2str(html string) string {
	src := string(html)

	//将HTML标签全转换成小写
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)

	//去除STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//去除SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	//去除连续的换行符
	re, _ = regexp.Compile("\\s{2,}")
	src = re.ReplaceAllString(src, "\n")

	return strings.TrimSpace(src)
}

func dateFormat(t time.Time, layout string) (datestring string) {
	datestring = t.Format(layout)
	return
}

func date(t time.Time, format string) (datestring string) {
	patterns := []string{
		// year
		"Y", "2006", // A full numeric representation of a year, 4 digits	Examples: 1999 or 2003
		"y", "06", //A two digit representation of a year	Examples: 99 or 03

		// month
		"m", "01", // Numeric representation of a month, with leading zeros	01 through 12
		"n", "1", // Numeric representation of a month, without leading zeros	1 through 12
		"M", "Jan", // A short textual representation of a month, three letters	Jan through Dec
		"F", "January", // A full textual representation of a month, such as January or March	January through December

		// day
		"d", "02", // Day of the month, 2 digits with leading zeros	01 to 31
		"j", "2", // Day of the month without leading zeros	1 to 31

		// week
		"D", "Mon", // A textual representation of a day, three letters	Mon through Sun
		"l", "Monday", // A full textual representation of the day of the week	Sunday through Saturday

		// time
		"g", "3", // 12-hour format of an hour without leading zeros	1 through 12
		"G", "15", // 24-hour format of an hour without leading zeros	0 through 23
		"h", "03", // 12-hour format of an hour with leading zeros	01 through 12
		"H", "15", // 24-hour format of an hour with leading zeros	00 through 23

		"a", "pm", // Lowercase Ante meridiem and Post meridiem	am or pm
		"A", "PM", // Uppercase Ante meridiem and Post meridiem	AM or PM

		"i", "04", // Minutes with leading zeros	00 to 59
		"s", "05", // Seconds, with leading zeros	00 through 59
	}
	replacer := strings.NewReplacer(patterns...)
	format = replacer.Replace(format)
	datestring = t.Format(format)
	return
}

// Compare is a quick and dirty comparison function. It will convert whatever you give it to strings and see if the two values are equal.
// Whitespace is trimmed. Used by the template parser as "eq"
func compare(a, b interface{}) (equal bool) {
	equal = false
	if strings.TrimSpace(fmt.Sprintf("%v", a)) == strings.TrimSpace(fmt.Sprintf("%v", b)) {
		equal = true
	}
	return
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
		Logger.Error("filepath.Walk() returned %v", err)
		return err
	}
	for k, v := range self.files {
		this.Templates[k] = template.Must(template.New("template" + k).Funcs(this.tplFuncMap).ParseFiles(v...))
	}
	return nil
}
