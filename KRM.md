TKO's KRM API
=============

The group-version is `tko.nephio.org/v1alpha1`. Supported kinds:

* `Template`
* `Site`
* `Deployment`
* `Plugin`

TKO's types are non-namespaced. They are associated with the cluster as a whole.

They are all in the `tko` category, so it's possible to access them collectively:

    scripts/kubectl-kind get tko

(We'll be using `scripts/kubectl-kind` in this guide. It's simply a shortcut to using the local
`kind-tko` configuration context.)

Individual types can be accessed by their names. Add `.tko` (or `.tko.nephio.org`) to avoid
ambiguity with other types. For example, to refer to TKO deployments rather than Kubernetes's
built-in deployments:

    scripts/kubectl-kind get deployment.tko

See the [examples](examples/kubernetes/).

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

The `tko kube` command can do this conversion for you:

    tko kube to demo/hello-world:v1.0.0
    > demo-002fhello-002dworld-003av1.0.0
    tko kube from demo-002fhello-002dworld-003av1.0.0
    > demo/hello-world:v1.0.0

Plugin IDs consist of the plugin type plus `|` plus the name, so when using a shell make sure to
wrap it in quotes so that your shell won't interpret the `|`:

    tko kube to 'validate|free5gc/smf'
    > validate-007cfree5gc-002fsmf

You can use `tko kube` inside a Bash expression, like so:

    kubectl get plugin $(tko kube to 'validate|free5gc/smf')

Note that you can select by the `metadata.name` field, which also accepts wildcards:

    scripts/kubectl-kind get template --field-selector=metadata.name=$(tko kube to nf/**)

Finally, note that deployment names have special semantics. When you create a new deployment, e.g.
with `kubectl create`, the KRM name is discarded, because the backend will generate a new random ID
for the name. So, simply write in a placeholder name, e.g. `placeholder-1`. However, note that every
time you create it, it will create a brand new deployment, even though you are using the same
placeholder name! This is similar to using
[`generateName`](https://kubernetes.io/docs/reference/using-api/api-concepts/#generated-values).

`kubectl apply` will *never* create new deployments and is *only* used for updates, in which case
you *do* need to specify the correct name and *also* the correct `metadata.resourceVersion`, which
TKO will use to verify that other changes haven't been made before yours. These are in fact the
proper KRM update semantics. Thus, the update workflow is to first retrieve (this will get you the
latest `resourceVersion`), then edit, then apply:

    scripts/kubectl-kind get deployment.tko/2d4fkjPFRYDKnvNGgKpuZs7kwGQ --output=yaml > d.yaml
    (edit d.yaml)
    scripts/kubectl-kind apply --filename=d.yaml

Or, do it all in one step using `kubectl edit`.

Metadata
--------

Metadata for templates, sites, and deployments are all represented as KRM labels.

However, note that KRM labels (both keys and values) have similar restrictions to names, thus they
must also be translated back and forth using `tko kube`.

Example:

```yaml
apiVersion: tko.nephio.org/v1alpha1
kind: Template

metadata:
  name: site-002fgdce-003av1.0.0 # site/gdce:v1.0.0
  labels:
    Site.cloud: GDC-002dE # GDC-E
    Site.region: chicago
    type: site
```

You can use label selectors to filter your results:

    scripts/kubectl-kind get template --selector=Site.cloud=$(tko kube to GDC-E)

Also note that the filter accepts wildcards:

    scripts/kubectl-kind get template --selector=Site.cloud=$(tko kube to G*)

Packages
--------

KRM packages for templates, sites, and deployments are in the `spec` under the `package.resources`
list, and are simply agnostic raw data. Example in YAML:

```yaml
apiVersion: tko.nephio.org/v1alpha1
kind: Template

metadata:
  name: k8s-002fhello-003av1.0.0 # k8s/hello:v1.0.0
  labels:
    m1: hello
    m2: world

spec:
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
            - name: nginx
              image: nginx:latest
```
