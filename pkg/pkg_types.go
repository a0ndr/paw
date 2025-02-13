package pkg

import (
	"github.com/BurntSushi/toml"
	"os"
	"path"
	"time"
)

type Package struct {
	//FQN              string
	Name             string
	Version          string
	Description      string
	Repository       string
	Checksum         string
	Dependencies     []string
	SoftDependencies []string
	Conflicts        []string
	BuiltAt          time.Time
}

type Build struct {
	Name        string `toml:"Name"`
	Version     string `toml:"Version"`
	Description string `toml:"Description"`
	Source      string `toml:"Source"`

	Dependencies     []string `toml:"Dependencies"`
	SoftDependencies []string `toml:"SoftDependencies"`
	Conflicts        []string `toml:"Conflicts"`

	Build string `toml:"Build"`

	PreInstall  string `toml:"PreInstall"`
	Install     string `toml:"Install"`
	PostInstall string `toml:"PostInstall"`

	Configure string `toml:"Configure"`

	PreRemove string `toml:"PreRemove"`
	Remove    string `toml:"Remove"`
}

type Meta struct {
	Name             string            `toml:"Name"`
	Version          string            `toml:"Version"`
	Description      string            `toml:"Description"`
	Checksums        map[string]string `toml:"Checksums"`
	Dependencies     []string          `toml:"Dependencies"`
	SoftDependencies []string          `toml:"SoftDependencies"`
	Conflicts        []string          `toml:"Conflicts"`
	BuiltAt          time.Time         `toml:"BuiltAt"`
}

type Entry struct {
	Name             string    `toml:"Name"`
	Version          string    `toml:"Version"`
	Description      string    `toml:"Description"`
	Checksum         string    `toml:"Checksum"`
	Dependencies     []string  `toml:"Dependencies"`
	SoftDependencies []string  `toml:"SoftDependencies"`
	Conflicts        []string  `toml:"Conflicts"`
	BuiltAt          time.Time `toml:"BuiltAt"`
}

type List struct {
	Packages map[string]*Entry `toml:"Packages"`
}

type DefList map[string]*Package

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

func (entry *Entry) ToDefinition(repo string) *Package {
	return &Package{
		//FQN:        	  fmt.Sprintf("%s-%s", entry.Name, entry.Version),
		Name:             entry.Name,
		Version:          entry.Version,
		Description:      entry.Description,
		Repository:       repo,
		Checksum:         entry.Checksum,
		Dependencies:     entry.Dependencies,
		SoftDependencies: entry.SoftDependencies,
		Conflicts:        entry.Conflicts,
		BuiltAt:          entry.BuiltAt,
	}
}

func (meta *Meta) ToEntry(checksum string) *Entry {
	return &Entry{
		Name:             meta.Name,
		Version:          meta.Version,
		Description:      meta.Description,
		Checksum:         checksum,
		Dependencies:     meta.Dependencies,
		SoftDependencies: meta.SoftDependencies,
		Conflicts:        meta.Conflicts,
		BuiltAt:          meta.BuiltAt,
	}
}
