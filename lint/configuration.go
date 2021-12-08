package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// ConfigurationFile contains a map for rule exclusions, and warnings, where the key is the
// rule name to be excluded or downgraded to a warning
type ConfigurationFile struct {
	Exclusions map[string]*ConfigurationRuleEntries `yaml:"exclusions"`
	Warnings   map[string]*ConfigurationRuleEntries `yaml:"warnings"`
}

type ConfigurationRuleEntries struct {
	Reason  string               `yaml:"reason,omitempty"`
	Entries []ConfigurationEntry `yaml:"entries,omitempty"`
}

// ConfigurationEntry will exist precisely once for every instance of a rule violation you wish
// exclude or downgrade to a warning. Each ConfigurationEntry will have to be an *exact* match
// to the combination of attributes set. Reason will not be evaluated, and is an opportunity for
// the author to explain why the exception, or downgrade to warning exists.
type ConfigurationEntry struct {
	Reason    string `yaml:"reason,omitempty"`
	Dashboard string `yaml:"dashboard,omitempty"`
	Panel     string `yaml:"panel,omitempty"`
	// This gets (un)marshalled as a string, because a 0 index is valid, but also the zero value of an int
	TargetIdx string `yaml:"targetIdx"`
}

func (cre *ConfigurationRuleEntries) AddEntry(e ConfigurationEntry) {
	cre.Entries = append(cre.Entries, e)
}

func (ce *ConfigurationEntry) IsMatch(r ResultContext) bool {
	ret := true
	if r.Dashboard != nil && ce.Dashboard != r.Dashboard.Title {
		ret = false
	}

	if r.Panel != nil && ce.Panel != r.Panel.Title {
		ret = false
	}

	if r.Target != nil && ce.TargetIdx != "" {
		idx, err := strconv.Atoi(ce.TargetIdx)
		if err == nil && idx != r.Target.Idx {
			ret = false
		}
	}

	return ret
}

func (cf *ConfigurationFile) Apply(res ResultContext) ResultContext {
	{
		exclusions, ok := cf.Exclusions[res.Rule.Name()]
		matched := false
		if exclusions != nil {
			for _, ce := range exclusions.Entries {
				if ce.IsMatch(res) {
					matched = true
				}
			}
			if len(exclusions.Entries) == 0 {
				matched = true
			}
		} else if ok {
			matched = true
		}
		if matched {
			res.Result.Severity = Exclude
			res.Result.Message = res.Result.Message + " (Excluded)"
		}
	}

	{
		warnings, ok := cf.Warnings[res.Rule.Name()]
		matched := false
		if warnings != nil {
			for _, ce := range warnings.Entries {
				if ce.IsMatch(res) {
					matched = true
				}
			}
			if len(warnings.Entries) == 0 {
				matched = true
			}
		} else if ok {
			matched = true
		}
		if matched {
			res.Result.Severity = Warning
		}
	}

	return res
}

func NewConfigurationFile() *ConfigurationFile {
	return &ConfigurationFile{
		Exclusions: map[string]*ConfigurationRuleEntries{},
		Warnings:   map[string]*ConfigurationRuleEntries{},
	}
}

func (cf *ConfigurationFile) Load(path string) error {
	lintFilePath := filepath.Join(path, ".lint")
	f, err := os.Open(lintFilePath)
	if err != nil && os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	if err = dec.Decode(cf); err != nil {
		return fmt.Errorf("could not unmarshal lint configuration %s: %w", lintFilePath, err)
	}
	return nil
}
