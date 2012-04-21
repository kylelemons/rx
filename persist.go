package main

import (
	"sync"

	"github.com/kylelemons/rx/repo"
)

var Repos repo.RepoMap
var scanOnce sync.Once

func Scan() (err error) {
	scanOnce.Do(func() {
		Repos, err = repo.Scan()
	})
	return
}
