package gounix

// TemplateEngine bracket wrapped string template with strings.Replacer.
type TemplateEngine interface {
	SetTemplate(template string) TemplateEngine
	AddParameter(name, value string) TemplateEngine
	Compile() string
}

// NewEngine create a new TemplateEngine.
func NewTemplate() TemplateEngine {
	engine := new(templateEngine)
	engine.params = make([]string, 0)
	return engine
}
