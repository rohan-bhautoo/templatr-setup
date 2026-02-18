package manifest

// Manifest represents the full .templatr.toml file structure.
type Manifest struct {
	Template  TemplateInfo      `toml:"template"`
	Runtimes  map[string]string `toml:"runtimes"`
	Packages  PackageConfig     `toml:"packages"`
	Env       []EnvVar          `toml:"env"`
	Config    []ConfigFile      `toml:"config"`
	PostSetup PostSetup         `toml:"post_setup"`
	Meta      Meta              `toml:"meta"`
}

// TemplateInfo identifies the template.
type TemplateInfo struct {
	Name     string `toml:"name"`
	Version  string `toml:"version"`
	Tier     string `toml:"tier"`
	Category string `toml:"category"`
	Slug     string `toml:"slug"`
}

// PackageConfig defines the package manager and install command.
type PackageConfig struct {
	Manager        string   `toml:"manager"`
	InstallCommand string   `toml:"install_command"`
	Global         []string `toml:"global,omitempty"`
}

// EnvVar defines a single environment variable for the .env file.
type EnvVar struct {
	Key         string `toml:"key"`
	Label       string `toml:"label"`
	Description string `toml:"description"`
	Default     string `toml:"default"`
	Required    bool   `toml:"required"`
	Type        string `toml:"type"` // text, url, email, secret, number, boolean
	DocsURL     string `toml:"docs_url,omitempty"`
}

// ConfigFile defines a configuration file to edit (e.g., site.ts).
type ConfigFile struct {
	File        string        `toml:"file"`
	Label       string        `toml:"label"`
	Description string        `toml:"description"`
	Fields      []ConfigField `toml:"fields"`
}

// ConfigField defines a single editable field within a config file.
type ConfigField struct {
	Path        string `toml:"path"`    // e.g. "siteConfig.name"
	Label       string `toml:"label"`
	Description string `toml:"description,omitempty"`
	Type        string `toml:"type"`    // text, url, email, number, boolean
	Default     string `toml:"default"`
}

// PostSetup defines commands and messages to show after setup completes.
type PostSetup struct {
	Commands []string `toml:"commands"`
	Message  string   `toml:"message"`
}

// Meta contains tool behavior configuration.
type Meta struct {
	MinToolVersion string `toml:"min_tool_version"`
	Docs           string `toml:"docs"`
}
