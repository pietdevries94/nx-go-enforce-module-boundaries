package analyzer

import "sync"

type projectFile struct {
	path string
	tags []string
}

type projectFileCacheFact struct {
	cache map[string]*projectFile
	m     sync.Mutex
}

type Boundaries struct {
	Allow          []string        `json:"allow"`
	DepConstraints []DepConstraint `json:"depConstraints"`
}

type DepConstraint struct {
	SourceTag                string   `json:"sourceTag"`
	OnlyDependOnLibsWithTags []string `json:"onlyDependOnLibsWithTags"`
	NotDependOnLibsWithTags  []string `json:"notDependOnLibsWithTags"`
}
