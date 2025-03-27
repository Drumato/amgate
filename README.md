# AlertManager Gate

**amgate** is a gateway that receives Alertmanager webhooks and triggers actions based on the alerts received.

## Configuration

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
    - matchers:
      - key: name
        op: "="
        value: "alert1"
      - labels:
          matchers:
          - key: severity
            op: "=~"
            value: "warning|critical"
      actor:
        kind: "rollout"
```