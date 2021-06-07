# kubectl-vpa

Tool to manage VPAs (vertical-pod-autoscaler) resources in a kubernetes-cluster
* Create, Compare, Change UpdateMode and Suggest limits

[![Go Report Card](https://goreportcard.com/badge/github.com/ninlil/kubectl-vpa)](https://goreportcard.com/report/github.com/ninlil/kubectl-vpa)

## Getting started
### Prerequisites

Install the [Vertical-Pod-Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) from the official [kubernetes/autoscaler](https://github.com/kubernetes/autoscaler)-project

### Recommended reading

[Vertical Pod Autoscaling: The Definitive Guide](https://povilasv.me/vertical-pod-autoscaling-the-definitive-guide/) by [Povilas Versockas](https://povilasv.me/)

## Install & Run

### To install:
```sh
go install github.com/ninlil/kubectl-vpa
```

### To run:
```sh
kubectl-vpa ...
```
or using the kubectl-plugin-behavior
```sh
kubectl vpa ...
```

### Help and options
All command have their own help. Example:
```sh
kubectl-vpa compare -h
```

## Create a VPA-resource

```sh
kubectl-vpa create foo/bar
```
This will find a Pod, Deployment, Daemonset, Statefulset or CronJob named 'bar' in the 'foo' namespace and output a document that is the VPA-resource of the found match.

## Change 'mode' of a VPA

```sh
kubectl-vpa mode initial -n foo bar1 bar2 bar3 database/mysql
```
This will set the UpdateMode to 'Initial' on VPA `bar1`, `bar2` & `bar3` in namespace `foo`, and on `mysql` in the `database` namespace

## Suggest limits (WIP)

```sh
kubectl-vpa suggest foo/bar
```
This will create snippets for us in a deployment (or other) resource describing requests if you do not want to use the 'recommender' module from the VPA

## Compare VPA with current requests

This will match current running pods and their current requests with matching VPA and output differences.

Sorting, filtering and head/tail of output is also available.


### Examples

Compare all VPA's in namespace `foo`
```s
kubectl-vpa compare -n foo
```
List all pods (including those without a matching VPA), with a sum-line
```s
kubectl-vpa compare -A -l -z
```

### The output

The following columns are printed:
* Namespace
* Name (name of pod)
* Mode (the UpdateMode, used by the `recommender`)
* Container
* Req-CPU (the request-cpu of the container in the current instance, in milli-units)
* VPA-CPU (the 'Target'-value of the matching VPA)
* CPU diff% (difference between the previous 2 values)
* Req-RAM (the request-memory of the container in the current instance, in M-units)
* VPA-RAM (the 'Target'-value of the matching VPA)
* Mem. diff% (difference between the previous 2 values)
* sum(Δ) (the sum of the 2 diff%-values)

The 'diff%' values will be positive when a container is requesting more that it probably needs, meaning a negative value is when it should probably request more than it's currently doing.

The 'Mode' column will display '---' onlines that don't match a VPA.