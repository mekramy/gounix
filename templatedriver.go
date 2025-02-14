package gounix

import "strings"

type templateEngine struct {
	template string
	params   []string
}

func (t *templateEngine) SetTemplate(template string) TemplateEngine {
	t.template = template
	return t
}

func (t *templateEngine) AddParameter(name, value string) TemplateEngine {
	t.params = append(t.params, "{"+name+"}", value)
	return t
}

func (t *templateEngine) Compile() string {
	return strings.NewReplacer(t.params...).Replace(t.template)
}
