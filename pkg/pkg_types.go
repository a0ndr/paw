package pkg

type Build struct {
	Name        string `toml:"Name"`
	Version     string `toml:"Version"`
	Description string `toml:"Description"`
	Source      string `toml:"Source"`

	Build string `toml:"Build"`

	PreInstall  string `toml:"PreInstall"`
	Install     string `toml:"Install"`
	PostInstall string `toml:"PostInstall"`

	Configure string `toml:"Configure"`
}

type Meta struct {
	Name        string            `toml:"Name"`
	Version     string            `toml:"Version"`
	Description string            `toml:"Description"`
	Checksums   map[string]string `toml:"Checksums"`
}

type List struct {
	Packages map[string]Meta `toml:"Packages"`
}
