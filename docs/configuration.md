# Configuration

## Environment variables

The following environment variables are used to configure the application:

- `AMGATE_NAMESPACE`: The namespace where the configmap(that is described below). Default is `amgate-system`.
- `AMGATE_CONFIGMAP_NAME`: The name of the configmap that contains the configuration. Default is `amgate-config`.

## ConfigMap

The configuration is done through a ConfigMap.
The default name is `amgate-config`.
The following is an example of a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: amgate-config
data:
  server: |
    host: "" # all interfaces
    port: 8080
  actions: |
    - name: k8s-rollout # build-in action
      matchers:
      - key: name
        op: "="
        value: "alert1"
      - labels:
          matchers:
          - key: severity
            op: "=~" # regex match by Go's regexp 
            value: "warning|critical"
      attrs:
        kind: Deployment
        namespace: apps
        name: myapp
        dry_run: false
```

### Matcher

A matcher is used to match the alert to the action.

supported operations:

- `=`: equal
- `!=`: not equal
- `=~`: regex match by Go's regexp

