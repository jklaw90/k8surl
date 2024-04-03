# k8surl

k8surl is a simple application to parse Kubernetes manifets and render templates and/or open links
it uses kubectls jsonpath lib but just adds some extra sugar to help you save those json templates somewhere with commands you can remember.


## Install

### Krew
```
kubectl krew install --manifest-url https://raw.githubusercontent.com/jklaw90/k8surl/main/plugins/k8surl.yaml
```

### From Source
Make sure to add a config file at `~/.k8surl.yaml`

```
go install github.com/jklaw90/k8surl@latest
```


## Examples

example `~/.k8surl.yaml` config file:

```yaml
pod:
  urls:
    - "https://example.com/dashboards/xyz?name={.metadata.name}&namespace={.metadata.namespace}"
  templates:
    - |
      Other Cool Dashboards:
      - "https://example.com/dashboards/xyz1?name={.metadata.name}&namespace={.metadata.namespace}"
commands:
  meta:
    short: open dashboards for namespaced things
    example: |
      kubectl k8surl meta get pods
      kubectl get pods -ojson | k8surl meta
    kinds: ["*"]
    urls:
      - "{range .items[*]}https://example.com/dashboards/xyz?name={.metadata.name}&namespace={.metadata.namespace}{end}"
    templates:
      - |
        {range .items[*]}
        Kind: {.kind}{"\t"}StartedTime:{.status.startTime}
        Name: {.metadata.name}
        {end}
  topology:
    kinds: [pod]
    templates:
      - |
        Topology Details:
          {.spec.topologySpreadConstraints[*].topologyKey}
```

example commands you can run from this file are
`k8surl meta` will open a dashboard and print out the template listed.

```
❯ kubectl get pod -ojson | k8surl meta
Kind: Pod	StartedTime:2023-05-11T17:33:28Z
Name: argocd-redis-845dd66445-676pn

Kind: Pod	StartedTime:2023-05-11T17:33:28Z
Name: argocd-notifications-controller-5bc5c7c44c-x4ll8

Kind: Pod	StartedTime:2023-05-11T17:33:27Z
Name: argocd-dex-server-76fb9b8c94-8cmww

Kind: Pod	StartedTime:2023-05-11T17:33:28Z
Name: argocd-applicationset-controller-596589bbc5-29wpf

Kind: Pod	StartedTime:2023-05-11T17:33:28Z
Name: argocd-server-577b775f6-2bxlx

Kind: Pod	StartedTime:2023-05-11T17:33:27Z
Name: argocd-repo-server-75bf74846d-tstlg

Kind: Pod	StartedTime:2023-05-11T17:34:23Z
Name: argocd-application-controller-0
```

If you define a type at the root like we did with `pod` it'll be your default configuration and you don't need more subcommands.

```
❯ kubectl get pod argocd-application-controller-0 -oyaml | k8surl
Other Cool Dashboards:
- "https://example.com/dashboards/xyz1?name=argocd-application-controller-0&namespace=argocd"
```
