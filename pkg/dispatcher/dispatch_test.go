package dispatcher

import (
	"testing"

	"github.com/Drumato/amgate/pkg/alertmanager"
	"github.com/Drumato/amgate/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestDispatchEventToActions(t *testing.T) {
	tests := []struct {
		name    string
		want    []DispatchResult
		cfg     *config.Config
		payload alertmanager.WebhookPayload
	}{
		{
			name: "empty config",
			cfg:  &config.Config{},
			payload: alertmanager.WebhookPayload{
				Alerts: []alertmanager.Alert{
					{
						Status: "firing",
					},
				},
			},
			want: []DispatchResult{},
		},
		{
			name: "empty payload",
			cfg: &config.Config{
				Actions: []config.ActionConfig{
					{
						Name: "test",
						Matchers: []config.MatcherConfig{
							{
								Key:   "status",
								Op:    "=",
								Value: "firing",
							},
						},
					},
				},
			},
			payload: alertmanager.WebhookPayload{},
			want:    []DispatchResult{},
		},
		{
			name: "empty alert",
			cfg: &config.Config{
				Actions: []config.ActionConfig{
					{
						Name: "test",
						Matchers: []config.MatcherConfig{
							{
								Key:   "status",
								Op:    "=",
								Value: "firing",
							},
						},
					},
				},
			},
			payload: alertmanager.WebhookPayload{
				Alerts: []alertmanager.Alert{},
			},
			want: []DispatchResult{},
		},
		{
			name: "alert matches to action",
			cfg: &config.Config{
				Actions: []config.ActionConfig{
					{
						Name: "test",
						Matchers: []config.MatcherConfig{
							{
								Key:   "status",
								Op:    "=",
								Value: "firing",
							},
						},
					},
				},
			},
			payload: alertmanager.WebhookPayload{
				Alerts: []alertmanager.Alert{
					{
						Status: "firing",
					},
				},
			},
			want: []DispatchResult{
				{
					ActionName: "test",
					Alert: DispatchAlert{
						Alert: alertmanager.Alert{
							Status: "firing",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DispatchEventToActions(tt.cfg, tt.payload)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_checkLabelMatcherMatchesToAlert(t *testing.T) {
	tests := []struct {
		name         string
		want         bool
		labelmatcher config.LabelMatcherConfig
		actualLabels map[string]string
	}{
		{
			name:         "empty label matcher",
			labelmatcher: config.LabelMatcherConfig{},
			actualLabels: map[string]string{},
			want:         true,
		},
		{
			name: "equal label matcher match",
			labelmatcher: config.LabelMatcherConfig{
				Matchers: []config.MatcherConfig{
					{
						Key:   "severity",
						Op:    "=",
						Value: "critical",
					},
				},
			},
			actualLabels: map[string]string{
				"severity": "critical",
			},
			want: true,
		},
		{
			name: "equal label matcher does not match",
			labelmatcher: config.LabelMatcherConfig{
				Matchers: []config.MatcherConfig{
					{
						Key:   "severity",
						Op:    "=",
						Value: "critical",
					},
				},
			},
			actualLabels: map[string]string{
				"severity": "warning",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkLabelMatcherMatchesToAlert(tt.actualLabels, tt.labelmatcher)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_checkMatcherMatchesToAlert_AlertField(t *testing.T) {
	tests := []struct {
		name         string
		want         bool
		matcher      config.MatcherConfig
		actualValues map[string]string
	}{
		{
			name:         "empty matcher",
			matcher:      config.MatcherConfig{},
			actualValues: map[string]string{},
			want:         false,
		},
		{
			name: "equal matcher match",
			matcher: config.MatcherConfig{
				Key:   "status",
				Op:    "=",
				Value: "firing",
			},
			actualValues: map[string]string{
				"status": "firing",
			},
			want: true,
		},
		{
			name: "equal matcher does not match",
			matcher: config.MatcherConfig{
				Key:   "status",
				Op:    "=",
				Value: "firing",
			},
			actualValues: map[string]string{
				"status": "resolved",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkMatcherMatchesToAlert(tt.actualValues, tt.matcher)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_checkMatchOperationBtwValues(t *testing.T) {
	tests := []struct {
		name    string
		want    bool
		matcher config.MatcherConfig
		actual  string
	}{
		{
			name:    "empty matcher",
			matcher: config.MatcherConfig{},
			actual:  "",
			want:    false,
		},
		{
			name: "equal matcher",
			matcher: config.MatcherConfig{
				Op:    "=",
				Value: "firing",
			},
			actual: "firing",
			want:   true,
		},
		{
			name: "not equal matcher",
			matcher: config.MatcherConfig{
				Op:    "!=",
				Value: "firing",
			},
			actual: "resolved",
			want:   true,
		},
		{
			name: "regex matcher",
			matcher: config.MatcherConfig{
				Op:    "=~",
				Value: "fir.*",
			},
			actual: "firing",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkMatcherOperationBtwValues(tt.matcher, tt.actual)
			assert.Equal(t, tt.want, got)
		})
	}
}
