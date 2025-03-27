package dispatcher

import (
	"regexp"

	"github.com/Drumato/amgate/pkg/alertmanager"
	"github.com/Drumato/amgate/pkg/config"
)

type DispatchResult struct {
	// ActionName is the name of the action that was dispatched.
	ActionName string
	Alert      DispatchAlert
}

type DispatchAlert struct {
	Alert alertmanager.Alert
	// the fields are flattened from the alertmanager.WebhookPayload struct
	Version           string
	GroupKey          string
	TruncatedAlerts   int
	Status            string
	Receiver          string
	GroupLabels       map[string]string
	CommonLabels      map[string]string
	CommonAnnotations map[string]string
	ExternalURL       string
}

// DispatchEventToActions dispatches the alertmanager webhook payload to the actions.
func DispatchEventToActions(
	cfg *config.Config,
	payload alertmanager.WebhookPayload,
) []DispatchResult {
	results := []DispatchResult{}

	for _, alert := range payload.Alerts {
		for _, action := range cfg.Actions {
			for _, matcher := range action.Matchers {
				if !checkLabelMatcherMatchesToAlert(alert.Labels, matcher.Labels) {
					continue
				}
				if !checkLabelMatcherMatchesToAlert(alert.Annotations, matcher.Annotations) {
					continue
				}
				if !checkLabelMatcherMatchesToAlert(payload.CommonLabels, matcher.CommonLabels) {
					continue
				}
				if !checkLabelMatcherMatchesToAlert(payload.CommonAnnotations, matcher.CommonAnnotations) {
					continue
				}

				actualValues := map[string]string{
					"status":       alert.Status,
					"startsAt":     alert.StartsAt,
					"endsAt":       alert.EndsAt,
					"generatorURL": alert.GeneratorURL,
					"fingerprint":  alert.Fingerprint,
				}
				if !checkMatcherMatchesToAlert(actualValues, matcher) {
					continue
				}

				results = append(results, DispatchResult{
					ActionName: action.Name,
					Alert: DispatchAlert{
						Alert:             alert,
						Version:           payload.Version,
						GroupKey:          payload.GroupKey,
						TruncatedAlerts:   payload.TruncatedAlerts,
						Status:            payload.Status,
						Receiver:          payload.Receiver,
						GroupLabels:       payload.GroupLabels,
						CommonLabels:      payload.CommonLabels,
						CommonAnnotations: payload.CommonAnnotations,
					},
				})
			}
		}
	}

	return results
}

func checkLabelMatcherMatchesToAlert(
	actualLabels map[string]string,
	labelmatcher config.LabelMatcherConfig,
) bool {
	for _, matcher := range labelmatcher.Matchers {
		if !checkMatcherMatchesToAlert(actualLabels, matcher) {
			return false
		}
	}
	return true
}

func checkMatcherMatchesToAlert(
	actualValuesMap map[string]string,
	matcher config.MatcherConfig,
) bool {
	actualValue, ok := actualValuesMap[matcher.Key]
	if !ok {
		// the expected value is not found.
		return false
	}

	if checkMatcherOperationBtwValues(matcher, actualValue) {
		return true
	}

	return false
}

func checkMatcherOperationBtwValues(
	matcher config.MatcherConfig,
	actualValue string) bool {
	switch matcher.Op {
	case "=":
		return actualValue == matcher.Value
	case "!=":
		return actualValue != matcher.Value
	case "=~":
		r, err := regexp.Compile(matcher.Value)
		if err != nil {
			// TODO: log
			return false
		}
		return r.Match([]byte(actualValue))
	default:
		return false
	}
}
