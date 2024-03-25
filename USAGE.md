TKO User Guide
==============

In this guide we'll assume that you followed the [installation guide](INSTALL.md) and have
the test scenario running.

### Accessing the CLI

The `tko` command is a straightforward entry point into TKO. It provides CLI access to all
the gRPC APIs. Let's start by making sure you can use `tko` to access the API server.

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
`tko` built on the host (via `scripts/build`) you can access the API server's exposed
port:

    scripts/tko-kind about

Use the CLI access that works for you for the rest of this guide.

### TUI dashboard

`tko` comes with a rich Terminal User Interface dashboard. To start it:

    tko dashboard

This dashboard is both keyboard and mouse friendly. Use arrow keys, tab, and escape key to
navigate. To quit press Q at the menu or click the Quit button. Also: pressing escape at the
menu or CTRL-C at any time.

The dashboard provides live views of all deployments, sites, templates, and plugins. You can
press enter or double click on individual cells to view more details, such as the full KRM
package. Here you can scroll with the keyboard or the mouse wheel. The escape key will exit the
details view.

### Web dashboard

The same dashboard is available over the web:

* For a native install, browse to: [http://localhost:50051](http://localhost:50051)
* For a Vagrant install: [http://localhost:60051](http://localhost:60051)
* For a Kubernetes Kind cluser install: [http://localhost:30051](http://localhost:30051)

### Working with entities

You can list, get, create, and delete all entities using the CLI. There's help for all
commands:

    tko deployment list --help

The `list` commands will return abbreviated results without the full KRM package data
(except `plugin list`, which will show full plugin information; plugins are not packages).
For example, to list all deployments:

    tko deployment list

Results are sorted by ID. Note that for deployments TKO uses "k-sortable" UUIDs, meaning
that the natural sort order of the IDs corresponds to their order of creation.

For very large amounts of results, TKO supports paging:

    tko deployment --offset=10 --max-count=10

All list commands have powerful filters to narrow results. You can combine filters.
Note that all queries happen at the backend and are scalable to millions of entities. A
few examples:

    tko deployment list --approved=false
    tko deployment list --site-id=lab/1
    tko deployment list --metadata=NetworkFunction.type=SMF

All ID and metadata filters support wildcards (again, handled by the backend). The `*`
wildcard stops at `/` and `:` boundaries, while the the `**` wildcard has no boundaries:

    tko deployment list --site-id=lab/*
    tko deployment list --template-id=nf/**:v1.0.0

To get all data for individual entities use `get` commands with exact IDs:

    tko template get topology/oran/cu:v1.0.0

Note that plugin IDs comprise both the type *and* the name:

    tko plugin get schedule kind

Use `delete` commands to delete individual entities:

    tko template delete topology/oran/cu:v1.0.0

You can combine listing and deleting via the `purge` commands (handled optimally by
the backend):

    tko deployment purge --site-id=lab/*

### Creating entities

Templates, sites, and plugins use `register` commands.

For templates, at the bare minimum you must provide an ID and a source for the KRM package, which
must be one or more YAML manifests. The source is a URL, which can be anything compatible with
[exturl](https://github.com/tliron/exturl/). It can a local file or directory (read recursively),
an archive file, an HTTP URL, and combinations.

Local directory example:

    tko template register demo/hello-world:v1.0.1 --url=examples/hello-world/

Over HTTP:

    tko template register demo/hello-world:v1.0.1 --url=https://raw.githubusercontent.com/nephio-experimental/tko/main/examples/hello-world/workload.yaml

You can also provide the KRM package via stdin:

    cat examples/hello-world/workload.yaml | tko template register demo/hello-world:v1.0.1 --stdin

If a previous entity with the ID already exists, it will be rewritten.

The package may already contain metadata via KRM. However, during registration it is
possible to add additional metadata or override package metadata:

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

    tko plugin register validate free5gc/smf examples/plugins/validate_free5gc_smf.py --trigger=free5gc.plugin.nephio.org,v1alpha1,SMF
    tko plugin register prepare namespace gcr.io/kpt-fn/set-namespace:v0.4.1 --executor=kpt --property=namespace=spec.namespace --trigger=workload.plugin,nephio.org,v1alpha1,Namespace

(Note that the executable file for the `command` executor must be accessible by the controllers.
In the Kubernetes cluster all plugins are executed in a special `tko-runner` pod.)

### Working with deployments

Creating deployments is a bit different from the other entities because an ID is generated for you.
Instead of a deployment ID, the first argument for `deployment create` is a template ID on which to base
the deployment:

    tko deployment create demo/hello-world:v1.0.0

It will return the generated ID, which you can capture into a shell variable:

    ID=$(tko deployment create demo/hello-world:v1.0.0)

You can provide a KRM package via a URL to merge into the template:

    tko deployment create demo/hello-world:v1.0.0 --url=examples/hello-world/

Via `create` you can also assign a site ID and a parent deployment ID.

    tko deployment create demo/hello-world:v1.0.0 --site=lab/1

The `approve` command is used to approve prepared deployments and works like `list` (and `purge`):

    tko deployment approve --template-id=nf/**:v1.0.0

We also provide CLI commands for modifying deployments. This involves starting a modification
and getting a token, and then either ending or cancelling the modification within a limited
time window. Example:

    # Create and get the deployment ID
    ID=$(tko deployment create topology/oran/cu:v1.0.0)
    # Extract the package into work directory and get a modification token
    M=$(tko deployment mod start "$ID" --url=work/)
    # Make a change (just an example)
    kpt fn eval --image=gcr.io/kpt-fn/set-namespace:v0.4.1 work/ -- namespace=network-function
    # End the modification
    tko deployment mod end "$M" --url=work/
