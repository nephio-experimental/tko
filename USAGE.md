TKO User Guide
==============

In this guide we'll assume that you've followed the [installation guide](INSTALL.md) and have
the test scenario running.

### Examining the test scenario results

The test scenario uses [a Kind scheduling plugin](examples/plugins/schedule_kind.py) to
provision Kubernetes clusters and schedule deployment packages on them, including Helm charts.

The scenario should result in two newly provisioned clusters, "edge1" and "edge2". If you're
running the TKO meta-scheduler natively then the clusters will be provisioned locally. Examples
for accessing them:

    kubectl get pods --all-namespaces --context=kind-edge1
    kubectl get pods --all-namespaces --context=kind-edge2

Note `kind-` as the prefix for the configuration context name.

If you've installed in a local Vagrant virtual machine the above would work if you're
in `vagrant ssh`. Or, you can run individual commands on the virtual machine
*from the host* like so:

    scripts/vagrant kubectl get pods --all-namespaces --context=kind-edge1

Finally, if you've installed on a local Kubernetes Kind cluster then the clusters will
be provisioned in the `tko-runner` pod, which is where the Kind scheduling plugin runs.
(This is Kind-in-Kind, using Docker-in-Docker.) To access `kubectl` there:

    scripts/kind-runner kubectl get pods --all-namespaces --context=kind-edge1

Initially you should see `upf` pods on the "edge1" cluster, but no workloads on the "edge2"
cluster. This is because the `smf` deployments were deliberately set to not auto-approve. You
can approve them like so (more information about using the CLI below):

    tko deployment approve

After a few seconds you should see `smf` pods appear on both "edge1" and "edger2" clusters.

### Accessing the CLI

The `tko` command is a straightforward client entry point into TKO. It provides CLI access to
all the backend APIs over gRPC. Let's start by making sure you can use `tko` to access the API
server.

If you've installed natively then this should just work:

    tko about

If you've installed in a local Vagrant virtual machine the above would work if you're
in `vagrant ssh`. Or, you can run individual commands on the virtual machine
*from the host* like so:

    scripts/vagrant tko about

A third option for Vagrant is to have `tko` built on the host (*not* just in the virtual
machine):

    scripts/build

And then you can access the API server inside the virtual machine via an exposed port. The
script for it:

    scripts/tko-vagrant about

Finally, if you've installed on a local Kubernetes Kind cluster then make sure you have
`tko` built on the host (via `scripts/build`) and then you can access the API server's
exposed port:

    scripts/tko-kind about

Use the CLI access that works for your installation for the rest of this guide.

### TUI dashboard

`tko` comes with a rich Terminal User Interface dashboard. To start it:

    tko dashboard

This dashboard is both keyboard and mouse friendly (for terminals that support mouse input).
Use tab, escape (or right mouse click), arrow keys (or mouse scroll wheel), and
page-up/page-down/home/end to navigate. To quit press Q or escape at the menu or click the Quit
button. You can also quit by pressing CTRL-C at any time.

The dashboard provides live views of all deployments, sites, templates, and plugins. You can
press enter or double click on individual table cells to view more details, such as the full KRM
package. Exit the detail view via escape or right mouse click.

### Web dashboard

A similar dashboard is available over the web. Note that though the TUI uses gRPC, the web GUI
uses HTTP. This is because gRPC client support in web browsers is currently too limited.

* For a native install, browse to: [http://localhost:50051](http://localhost:50051)
* For a Vagrant install: [http://localhost:60051](http://localhost:60051)
* For a Kubernetes Kind cluster install: [http://localhost:30051](http://localhost:30051)

### Examining entities

You can list and get all entities using the `tko` CLI. There's help for all commands, for
example:

    tko deployment list --help

The `list` commands will return abbreviated results without the full KRM package data
(except `plugin list`, which will show full plugin information; plugins are not packages).
For example, to list all deployments:

    tko deployment list

Results are sorted by ID. Note that for deployments TKO uses "K-sortable" UUIDs, meaning
that the natural string sort order of the IDs corresponds to their order of creation.

`tko` defaults to YAML representations, but many other formats are supported via
the `--format` argument: "yaml", "json", "xjson", "xml", "cbor", "messagepack", and "go".
Examples:

    tko deployment list --format=json
    tko deployment list --format=go

(Note that the "cbor" and "messagepack" formats are binary and may not be represented
properly in terminals.)

For handling large amounts of results, `tko` supports paging:

    tko deployment list --offset=10 --max-count=10

All list commands have powerful filters to narrow results, and you can combine filters.
See `--help` for available filters. Note that all queries happen at the backend and are
scalable to millions of entities. A few examples:

    tko deployment list --approved=false
    tko deployment list --site-id=lab/1
    tko deployment list --metadata=NetworkFunction.type=SMF

All ID and metadata filters support wildcards (again, handled by the backend). The `*`
wildcard stops at `/` and `:` boundaries, while the the `**` wildcard has no boundaries:

    tko deployment list --site-id=lab/*
    tko deployment list --template-id=nf/**:v1.0.0

To get the package for individual entities use `get` commands with exact IDs:

    tko template get topology/oran/cu:v1.0.0

Note that plugin IDs comprise both the type *and* the name for `plugin get`:

    tko plugin get schedule kind

### Registering entities

For templates, at the bare minimum you must provide an ID and a source for the KRM package, which
must be one or more YAML manifests. The source is a URL, which can be anything compatible with
[exturl](https://github.com/tliron/exturl/). It can a local file or directory (read recursively),
an archive (tarball or zip), an HTTP URL, a git URL, and combinations.

Local directory example:

    tko template register demo/hello-world:v1.0.1 --url=examples/hello-world/

HTTP example:

    tko template register demo/hello-world:v1.0.1 --url=https://raw.githubusercontent.com/nephio-experimental/tko/main/examples/hello-world/workload.yaml

You can also provide the KRM package via stdin:

    cat examples/hello-world/workload.yaml | tko template register demo/hello-world:v1.0.1 --stdin

If a previous entity with the ID already exists then it will be rewritten, though note that
existing relationships (template to deployments, deployments to site, etc.) will be maintained.
In production use cases it's more likely to register a new entity with a different ID, for example
by appending version information to the ID as we did in our examples.

The package may already contain metadata via KRM (see [packages reference](PACKAGES.md#metadatanephioorg)).
However, during registration it is possible to add additional metadata or override package metadata:

    tko template register demo/hello-world:v1.0.1 --metadata=hello=world --url=examples/hello-world/

Registering a site is similar except that you can optionally base it on a template, in which case
any additional package you provide will be merged into the template. It is also possible to register
a site with no package data. A few examples:

    # Empty
    tko site register lab/2

    # Based on template
    tko site register india/bangalore/south-102 site/gdce:v1.0.0

    # Based on package
    tko site register lab/3 --url=examples/lab-site/

Registering plugins doesn't involve a package, but instead relies on arguments and properties
depending on the executor (the default executor is `command`). Triggers are provided as
comma-separated KRM GVKs. Two examples:

    # "command" executor (the default)
    tko plugin register validate free5gc/smf examples/plugins/validate_free5gc_smf.py --trigger=free5gc.plugin.nephio.org,v1alpha1,SMF

    # "kpt" executor
    tko plugin register prepare namespace gcr.io/kpt-fn/set-namespace:v0.4.1 --executor=kpt --property=namespace=spec.namespace --trigger=workload.plugin,nephio.org,v1alpha1,Namespace

(Note that the executable file for the `command` executor must be accessible by the controllers.
In the Kubernetes cluster all plugins are executed in a special `tko-runner` pod.)

### Deleting entities

Use `delete` commands to delete individual entities by their exact IDs:

    tko template delete topology/oran/cu:v1.0.0

You can combine listing and deleting via the `purge` commands (again handled optimally
by the backend):

    tko deployment purge --site-id=lab/*

### Working with deployments

Creating deployments is a bit different from the other entities because an ID is generated for you.
Instead of a deployment ID, the first argument for `deployment create` is a template ID on which to base
the deployment:

    tko deployment create demo/hello-world:v1.0.0

It will return the generated ID, which you can capture into a shell variable:

    ID=$(tko deployment create demo/hello-world:v1.0.0)

You can provide a KRM package via a URL to merge into the template:

    tko deployment create demo/hello-world:v1.0.0 --url=examples/hello-world/

You can also assign a site ID and/or a parent deployment ID:

    tko deployment create demo/hello-world:v1.0.0 --site=lab/1

The `approve` command is used to approve prepared and unpparoved deployments and supports most of the
same filters as `list` (and `purge`):

    tko deployment approve --template-id=nf/**:v1.0.0

We also provide CLI commands for modifying deployments. This involves starting a modification
and getting a token, and then either ending or cancelling the modification within a limited
time window. During that window other clients cannot modify the deployment. Example:

    # Create and get the deployment ID
    ID=$(tko deployment create demo/hello-world:v1.0.0)

    # Extract the package into work directory and get a modification token
    mkdir --parents /tmp/mywork
    M=$(tko deployment mod start "$ID" --url=/tmp/mywork/)

    # Make a change to the package (just an example!)
    kpt fn eval --image=gcr.io/kpt-fn/set-namespace:v0.4.1 /tmp/mywork/ -- namespace=mynamespace

    # End the modification and send the changed package data
    tko deployment mod end "$M" --url=/tmp/mywork/

### Using the KRM API

If you've installed TKO in a Kubernetes cluster then you can use its aggregated KRM API as an
alternative to gRPC. Essentially, most of what you can do with the `tko` CLI can be done via
`kubectl` (or any other Kubernetes client) instead.

See the [KRM guide](KRM.md) for more information. Here we'll provide just a few quick usage
examples.

An equivalent of `tko template get`:

    scripts/kubectl-kind get template/$(tko kube to demo/hello-world:v1.0.0) --output=yaml

Note the use of `tko kube` to convert IDs, explained in the KRM guide.

If you want to extract just the package you can use [yq](https://github.com/mikefarah/yq):

    scripts/kubectl-kind get template/$(tko kube to demo/hello-world:v1.0.0) --output=yaml | yq .spec.package.resources

Another way to do this is to get the package in JSON and then convert it to YAML (with yq):

    scripts/kubectl-kind get template/$(tko kube to demo/hello-world:v1.0.0) --output=jsonpath={.spec.package.resources} | yq --prettyPrint

Here's an equivalent of `tko template list` with wildcards:

    scripts/kubectl-kind get template --field-selector=metadata.name=$(tko kube to nf/**)

