[![CI](https://github.com/Drumato/amgate/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/Drumato/amgate/actions/workflows/ci.yaml)
[![Release](https://github.com/Drumato/amgate/actions/workflows/build_image.yaml/badge.svg)](https://github.com/Drumato/amgate/actions/workflows/build_image.yaml)

# AlertManager Gate

**amgate** is a gateway that receives Alertmanager webhooks and triggers actions based on the alerts received.

## Roadmap

- [x] Base Configuration
- [ ] Base Actions
    - [x] Kubernetes Rollout
    - [ ] Modify K8s manifests and push them to a Git repository
- [ ] Helm Chart

## Documents

- [Configuration](docs/configuration.md)
- [Actions](docs/actions.md)
- [Framework](docs/framework.md)
