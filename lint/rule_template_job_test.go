package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJobDatasource(t *testing.T) {
	linter := NewTemplateJobRule()

	for _, tc := range []struct {
		result    Result
		dashboard Dashboard
	}{
		// Non-promtheus dashboards shouldn't fail.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		// Missing job template.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' is missing the job template",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
					},
				},
			},
		},
		// Wrong datasource.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' job template should use datasource '$datasource'",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "job",
							Datasource: "foo",
						},
					},
				},
			},
		},
		// Wrong type.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' job template should be a Prometheus query",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "job",
							Datasource: "$datasource",
							Type:       "bar",
						},
					},
				},
			},
		},
		// Wrong job label.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' job template should be a labelled 'job'",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "bar",
						},
					},
				},
			},
		},
		// What success looks like.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "job",
							Multi:      true,
							AllValue:   ".+",
						},
						{
							Name:       "instance",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "instance",
							Multi:      true,
							AllValue:   ".+",
						},
					},
				},
			},
		},
	} {
		require.Equal(t, tc.result, linter.LintDashboard(tc.dashboard))
	}
}
