package cmd

type Config struct {
	KindAndTemplates KindAndTemplates        `mapstructure:",remain"`
	ExtraCommands    map[string]ExtraCommand `yaml:"extraCommands"`
}

type KindAndTemplates map[string]Tmpl

type ExtraCommand struct {
	Tmpl  `mapstructure:",squash"`
	Kinds []string `yaml:"kinds"`
}

type Tmpl struct {
	Urls      []string `yaml:"urls"`
	Templates []string `yaml:"templates"`
}
