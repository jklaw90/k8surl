package cmd

type Config struct {
	KindAndTemplates KindAndTemplates   `mapstructure:",remain"`
	Commands         map[string]Command `yaml:"commands"`
}

type KindAndTemplates map[string]Tmpl

type Command struct {
	Short   *string `yaml:"short"`
	Example *string `yaml:"example"`
	Tmpl    `mapstructure:",squash"`
	Kinds   []string `yaml:"kinds"`
}

type Tmpl struct {
	Urls      []string `yaml:"urls"`
	Templates []string `yaml:"templates"`
}
