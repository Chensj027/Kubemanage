---
name: Kubemanage deployment target
description: node234 is intended formal K8s deployment node; node191 is current campus-access dev node.
type: project
created: 2026-07-04
updated: 2026-07-04
---

Why: Future deployment work should target the intended production node and avoid assuming node191 is final.

How to apply: For formal manifests, pin workloads with `nodeSelector` `kubernetes.io/hostname: node234`; use node234 InternalIP `10.90.1.234` for NodePort access. Treat `10.90.1.191` as the current development node reachable on the campus network.
