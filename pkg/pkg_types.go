package pkg

import (
	"github.com/BurntSushi/toml"
	"os"
	"path"
)

type Definition struct {
	//FQN         string
	Name        string
	Version     string
	Description string
	Repository  string
	Checksum    string
}

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

	PreRemove string `toml:"PreRemove"`
	Remove    string `toml:"Remove"`
}

type Meta struct {
	Name        string            `toml:"Name"`
	Version     string            `toml:"Version"`
	Description string            `toml:"Description"`
	Checksums   map[string]string `toml:"Checksums"`
}

type Entry struct {
	Name        string `toml:"Name"`
	Version     string `toml:"Version"`
	Description string `toml:"Description"`
	Checksum    string `toml:"Checksum"`
}

type List struct {
	Packages map[string]*Entry `toml:"Packages"`
}

type DefList map[string]*Definition

type Cache struct {
	Packages *DefList `toml:"Packages"`
}

func (cache *Cache) Load() error {
	cache.Packages = &DefList{}

	fileContent, err := os.ReadFile(path.Join(Cfg.DataDir, "cache.toml"))
	if err != nil {
		return err
	}

	_, err = toml.Decode(string(fileContent), cache)
	if err != nil {
		return err
	}

	return nil
}

func (entry *Entry) ToDefinition(repo string) *Definition {
	return &Definition{
		//FQN:         fmt.Sprintf("%s-%s", entry.Name, entry.Version),
		Name:        entry.Name,
		Version:     entry.Version,
		Description: entry.Description,
		Repository:  repo,
		Checksum:    entry.Checksum,
	}
}

func (meta *Meta) ToEntry(checksum string) *Entry {
	return &Entry{
		Name:        meta.Name,
		Version:     meta.Version,
		Description: meta.Description,
		Checksum:    checksum,
	}
}
