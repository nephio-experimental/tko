TKO's KRM API
=============

TKO's types are non-namespaced. They are associated with the cluster as a whole.

They are all in the `tko` category, so it's possible to access them collectively:

    kubectl get tko

Individual types can be accessed by their names. Add `.tko` (or `.tko.nephio.org`) to avoid
ambiguity with other types. For example, to refer to TKO deployments rather than Kubernetes's
built-in deployments:

    kubectl get deployment.tko

Names
-----

Kubernetes resource names have restrictions. They must be DNS-compatible because in some cases
(e.g. pods and services) they indeed are used as DNS names. Unfortunately, this restriction applies
to all names, even if they are never associated with DNS. Essentially, names must be alphanumeric
and the only allowed punctuation is dashes and dots.

To get around this limitation, TKO escapes forbidden characters using the slash plus the Unicode
number in hex. So, this ID:

    demo/hello-world:v1.0.0

becomes this name:

    demo-002fhello-002dworld-003av1.0.0

The `tko` command can do this conversion for you:

    tko name to demo/hello-world:v1.0.0
    > demo-002fhello-002dworld-003av1.0.0
    tko name from demo-002fhello-002dworld-003av1.0.0
    > demo/hello-world:v1.0.0

Plugin IDs consist of the plugin type plus `|` plus the name, so when using a shell make sure to
wrap the it in quotes so that your shell won't interpret the `|`:

    tko name to 'validate|free5gc/smf'
    > validate-007cfree5gc-002fsmf

Template Type
-------------

We can create/update a template like so:

```yaml
apiVersion: tko.nephio.org/v1alpha1
kind: Template
metadata:
  name: k8s-002fhello-003av1.0.0 # k8s/hello:v1.0.0
spec:
  metadata:
    type: demo
  package:
    resources:
    - apiVersion: v1
      kind: Namespace
      metadata:
        name: hello-world
    - apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: hello-world
        namespace: hello-world
      spec:
        replicas: 1
        selector:
          matchLabels:
            app.kubernetes.io/name: hello-world
        template:
          metadata:
            labels:
              app.kubernetes.io/name: hello-world
          spec:
            containers:
            - image: nginx:latest
              name: nginx
```

Site Type
---------

Deployment Type
---------------

Plugin Type
-----------
