package action

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/Drumato/amgate/pkg/dispatcher"
	"github.com/cockroachdb/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sRolloutAction struct {
	logger    *slog.Logger
	k8sClient client.Client
}

func (a *K8sRolloutAction) Name() string {
	return "k8s-rollout"
}

func (a *K8sRolloutAction) Run(ctx context.Context, result dispatcher.DispatchResult) error {
	cfg := a.collectConfig(result.Attrs)

	// start rollout like `kubectl rollout restart`
	// https://github.com/kubernetes/kubectl/blob/fd89c3d1570b30935474a96cf42677d89faa2482/pkg/polymorphichelpers/objectrestarter.go#L32

	var patch client.Patch
	var targetObject client.Object

	switch cfg.Kind {
	case "Deployment":
		deployment := appsv1.Deployment{}
		if err := a.k8sClient.Get(ctx, types.NamespacedName{
			Namespace: cfg.Namespace,
			Name:      cfg.Name,
		}, &deployment); err != nil {
			return errors.WithStack(err)
		}
		patch = client.MergeFrom(deployment.DeepCopy())

		if deployment.Spec.Template.Labels == nil {
			deployment.Spec.Template.Labels = map[string]string{}
		}
		deployment.Spec.Template.Labels["amgate.drumato.com/rollout"] = "true"
		deployment.Spec.Template.Labels["amgate.drumato.com/restart-at"] = time.Now().Format(time.RFC3339)
		targetObject = &deployment
	case "StatefulSet":
		statefulSet := appsv1.StatefulSet{}
		if err := a.k8sClient.Get(ctx, types.NamespacedName{
			Namespace: cfg.Namespace,
			Name:      cfg.Name,
		}, &statefulSet); err != nil {
			return errors.WithStack(err)
		}
		patch = client.MergeFrom(statefulSet.DeepCopy())

		if statefulSet.Spec.Template.Labels == nil {
			statefulSet.Spec.Template.Labels = map[string]string{}
		}
		statefulSet.Spec.Template.Labels["amgate.drumato.com/rollout"] = "true"
		statefulSet.Spec.Template.Labels["amgate.drumato.com/restart-at"] = time.Now().Format(time.RFC3339)

		targetObject = &statefulSet
	case "DaemonSet":
		daemonSet := appsv1.DaemonSet{}
		if err := a.k8sClient.Get(ctx, types.NamespacedName{
			Namespace: cfg.Namespace,
			Name:      cfg.Name,
		}, &daemonSet); err != nil {
			return errors.WithStack(err)
		}
		patch = client.MergeFrom(daemonSet.DeepCopy())

		if daemonSet.Spec.Template.Labels == nil {
			daemonSet.Spec.Template.Labels = map[string]string{}
		}
		daemonSet.Spec.Template.Labels["amgate.drumato.com/rollout"] = "true"
		daemonSet.Spec.Template.Labels["amgate.drumato.com/restart-at"] = time.Now().Format(time.RFC3339)

		targetObject = &daemonSet
	}

	if !cfg.DryRun {
		if err := a.k8sClient.Patch(ctx, targetObject, patch); err != nil {
			return errors.WithStack(err)
		}
	} else {
		// dry-run
		a.logger.Info("dry-run", slog.String("kind", cfg.Kind), slog.String("namespace", cfg.Namespace), slog.String("name", cfg.Name))
	}

	return nil
}

type K8sRolloutConfig struct {
	DryRun bool

	Kind      string
	Namespace string
	Name      string
}

func (a *K8sRolloutAction) collectConfig(attrs map[string]string) K8sRolloutConfig {
	cfg := K8sRolloutConfig{}
	cfg.Kind = attrs["kind"]
	cfg.Namespace = attrs["namespace"]
	cfg.Name = attrs["name"]
	dryRunCfg, ok := attrs["dry_run"]
	if ok {
		dryRun, err := strconv.ParseBool(dryRunCfg)
		if err != nil {
			cfg.DryRun = dryRun
		}
	}

	return cfg
}

func NewK8sRolloutAction(
	logger *slog.Logger,
	k8sClient client.Client,
) Action {
	actionLogger := logger.With(slog.String("action", "k8s-rollout"))
	return &K8sRolloutAction{
		logger:    actionLogger,
		k8sClient: k8sClient,
	}
}
