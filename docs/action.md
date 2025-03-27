# Action Framework

The Action Framework is a powerful tool that allows you to create custom actions that can be executed by the server. Actions can be used to automate tasks, interact with external services.

## Action Interface

every action implements the `Action` interface.

```go
type Action interface {
    Name() string
	Run(
		ctx context.Context,
		result dispatcher.DispatchResult,
	) error
}
```

`result` has attrs that can be used to pass data into the action.

## Built-in Actions

some actions are built-in and can be used out of the box.

### K8s Rollout

The `k8s-rollout` action can be used to rollout a new version of a Kubernetes deployment/daemonset/statefulset.

the attributes of this action are:

```yaml
kind: "Deployment" # Deployment, DaemonSet, StatefulSet
namespace: "default"
name: "my-deployment"
dry_run: false # true to debug
```
