package gounix

import "strings"

type templateEngine struct {
	template string
	params   []string
}

func (engine *templateEngine) SetTemplate(template string) TemplateEngine {
	engine.template = template
	return engine
}

func (engine *templateEngine) AddParameter(name, value string) TemplateEngine {
	engine.params = append(engine.params, "{"+name+"}", value)
	return engine
}

func (engine *templateEngine) Compile() string {
	return strings.NewReplacer(engine.params...).Replace(engine.template)
}
