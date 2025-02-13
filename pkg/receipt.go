package pkg

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"path"
	"sort"
	"time"
)

type Operation struct {
	Package   *Package
	Operation string
}

type Receipt struct {
	Date       time.Time
	Operations []Operation
	Completed  bool
}

func CreateReceipt() Receipt {
	return Receipt{
		Date:       time.Now(),
		Operations: []Operation{},
	}
}

func (r *Receipt) AddOperation(pkg *Package, op string) {
	r.Operations = append(r.Operations, Operation{
		Package:   pkg,
		Operation: op,
	})
}

func (r *Receipt) HasPackage(pkg *Package) bool {
	for _, op := range r.Operations {
		if op.Package.Name == pkg.Name {
			return true
		}
	}

	return false
}

func (r *Receipt) Flush() error {
	p := path.Join(Cfg.DataDir, "receipts", r.Date.Format(time.RFC3339)+".toml")
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	encoder := toml.NewEncoder(f)
	return encoder.Encode(r)
}

func (r *Receipt) Delete() error {
	p := path.Join(Cfg.DataDir, "Receipts", r.Date.Format(time.RFC3339)+".toml")
	return os.Remove(p)
}

func (cfg *Config) LoadReceipts() error {
	receiptFiles, err := os.ReadDir(path.Join(cfg.DataDir, "receipts"))
	if err != nil {
		fmt.Println(1)
		return err
	}

	fis := make([]os.FileInfo, 0, len(receiptFiles))

	for _, file := range receiptFiles {
		fi, err := file.Info()
		if err != nil {
			fmt.Println(2)
			return err
		}
		fis = append(fis, fi)
	}

	sort.Slice(fis, func(i, j int) bool {
		return fis[i].ModTime().After(fis[j].ModTime())
	})

	for _, file := range fis {
		r := &Receipt{}
		if _, err := toml.DecodeFile(path.Join(cfg.DataDir, "receipts", file.Name()), r); err != nil {
			fmt.Println(3)
			return err
		}
		cfg.Receipts = append(cfg.Receipts, r)
	}

	return nil
}

func (cfg *Config) FindNewestUncompletedReceipt() *Receipt {
	for _, receipt := range cfg.Receipts {
		if !receipt.Completed {
			return receipt
		}
	}

	return nil
}
