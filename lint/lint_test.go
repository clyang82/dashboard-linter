package lint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestRule struct {
	name string
}

func (r *TestRule) Description() string {
	return "Test Rule"
}

func (r *TestRule) Name() string {
	return r.name
}

func appendConfigExclude(t *testing.T, rule string, dashboard string, panel string, targetIdx string, config *ConfigurationFile) {
	t.Helper()

	entries := config.Exclusions[rule]
	if entries == nil {
		entries = &ConfigurationRuleEntries{}
	}

	if dashboard != "" || panel != "" || targetIdx != "" {
		entries.Entries = append(entries.Entries, ConfigurationEntry{
			Dashboard: dashboard,
			Panel:     panel,
			TargetIdx: targetIdx,
		})
	}
	config.Exclusions[rule] = entries
}

func appendConfigWarning(t *testing.T, rule string, dashboard string, panel string, targetIdx string, config *ConfigurationFile) {
	t.Helper()

	entries := config.Warnings[rule]
	if entries == nil {
		entries = &ConfigurationRuleEntries{}
	}

	if dashboard != "" || panel != "" || targetIdx != "" {
		entries.Entries = append(entries.Entries, ConfigurationEntry{
			Dashboard: dashboard,
			Panel:     panel,
			TargetIdx: targetIdx,
		})
	}
	config.Warnings[rule] = entries
}

func newResultContext(t *testing.T, rule string, dashboard string, panel string, targetIdx string, result Severity) ResultContext {
	ret := ResultContext{
		Result: Result{Severity: result, Message: "foo"},
	}
	if rule != "" {
		ret.Rule = &TestRule{name: rule}
	}
	if dashboard != "" {
		ret.Dashboard = &Dashboard{Title: dashboard}
	}
	if panel != "" {
		ret.Panel = &Panel{Title: panel}
	}
	if targetIdx != "" {
		idx, err := strconv.Atoi(targetIdx)
		if err == nil {
			ret.Target = &Target{Idx: idx}
		}
	}
	return ret
}

func TestResultSet(t *testing.T) {
	t.Run("MaximumSeverity", func(t *testing.T) {
		r := ResultSet{
			results: []ResultContext{
				{Result: Result{Severity: Success}},
				{Result: Result{Severity: Warning}},
				{Result: Result{Severity: Error}},
			},
		}

		require.Equal(t, r.MaximumSeverity(), Error)
	})

	t.Run("ByRule", func(t *testing.T) {
		r := ResultSet{
			results: []ResultContext{
				newResultContext(t, "rule1", "", "", "", Success),
				newResultContext(t, "rule2", "", "", "", Success),
			},
		}

		byRule := r.ByRule()

		require.Len(t, byRule, 2)
		require.Contains(t, byRule, "rule1")
		require.Contains(t, byRule, "rule2")
		require.Len(t, byRule["rule1"], 1)
		require.Len(t, byRule["rule2"], 1)
	})

	t.Run("Honors Configuration given config present before results added", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "", "", "", c)

		r := ResultSet{}
		r.Configure(c)
		r.AddResult(newResultContext(t, "rule1", "", "", "", Error))

		require.Equal(t, Exclude, r.MaximumSeverity())
		require.Equal(t, Exclude, r.ByRule()["rule1"][0].Result.Severity)
	})

	t.Run("Honors Configuration given config added after results added", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "", "", "", c)

		r := ResultSet{}
		r.AddResult(newResultContext(t, "rule1", "", "", "", Error))
		r.Configure(c)

		require.Equal(t, Exclude, r.MaximumSeverity())
		require.Equal(t, Exclude, r.ByRule()["rule1"][0].Result.Severity)
	})
}

func TestConfiguration(t *testing.T) {
	t.Run("Excludes Rule", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "", "", "", c)

		r1 := newResultContext(t, "rule1", "", "", "", Error)
		r2 := newResultContext(t, "rule2", "", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Rule", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigWarning(t, "rule1", "", "", "", c)

		r1 := newResultContext(t, "rule1", "", "", "", Error)
		r2 := newResultContext(t, "rule2", "", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Excludes More Specific Config", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "", "", "", c)
		appendConfigExclude(t, "rule1", "dash1", "", "", c)

		r1 := newResultContext(t, "rule1", "dash1", "", "", Error)
		r2 := newResultContext(t, "rule1", "dash2", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Excludes all when rule defined but entries empty", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "", "", "", c)

		r1 := newResultContext(t, "rule1", "dash1", "panel1", "0", Error)
		r2 := newResultContext(t, "rule1", "dash1", "panel1", "1", Error)

		rs := []ResultContext{r1, r2}
		for _, r := range rs {
			rc := c.Apply(r)
			require.Equal(t, Exclude, rc.Result.Severity)
		}
	})

	// Dashboards
	t.Run("Excludes Dashboard", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "dash1", "", "", c)

		r1 := newResultContext(t, "rule1", "dash1", "", "", Error)
		r2 := newResultContext(t, "rule1", "dash2", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Dashboard", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigWarning(t, "rule1", "dash1", "", "", c)

		r1 := newResultContext(t, "rule1", "dash1", "", "", Error)
		r2 := newResultContext(t, "rule1", "dash2", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	// Panels
	t.Run("Excludes Panels", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "dash1", "panel1", "", c)

		r1 := newResultContext(t, "rule1", "dash1", "panel1", "", Error)
		r2 := newResultContext(t, "rule1", "dash1", "panel2", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Panels", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigWarning(t, "rule1", "dash1", "panel1", "", c)

		r1 := newResultContext(t, "rule1", "dash1", "panel1", "", Error)
		r2 := newResultContext(t, "rule1", "dash1", "panel2", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	// Targets
	t.Run("Excludes Targets", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigExclude(t, "rule1", "dash1", "panel1", "0", c)

		r1 := newResultContext(t, "rule1", "dash1", "panel1", "0", Error)
		r2 := newResultContext(t, "rule1", "dash1", "panel1", "1", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Targets", func(t *testing.T) {
		c := NewConfigurationFile()
		appendConfigWarning(t, "rule1", "dash1", "panel1", "0", c)

		r1 := newResultContext(t, "rule1", "dash1", "panel1", "0", Error)
		r2 := newResultContext(t, "rule1", "dash1", "panel1", "1", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})
}
