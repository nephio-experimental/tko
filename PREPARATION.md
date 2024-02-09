TKO Preparation
===============

The Loop
--------

Preparation works as an iterative loop over all unprepared deployment packages.
Within each package, specific resources will be prepared (not all resources need to be
prepared). It's common that a single iteration may not be enough, as each iteration may
add new resources that would also need to prepared. It's also possible that preparation
may cause other backend changes, for example the creation of new child deployments.

The preparation status per resource is set via the `nephio.org/prepared` annotation.

Once all preparable resources have been prepared, the Preparer marks the whole package as
prepared by setting that annotation on a single resource of GVK
`deployment.nephio.org/v1alpha1`, `Deployment`. The Prepaper will automatically add
this resource to a deployment package if it doesn't already exist. This annotation
bubbles up to the deployment entity on the backend, for example as a column in a SQL
table, making it trivial to query deployments by preparation state.

Note that a prepared package can be forced to re-prepare if the annotation is removed
or set to "false".

Plugins
-------

Preparation plugins are triggered by the presence of specific preparable GVK. Other
resources are ignored (they are assumed to be already prepared).

Note that this is simply an optimization! We could potentially run *all* available
plugins, they would simply do nothing if they don't find supported resources.

Simply registering a preparation plugin for a GVK will automatically mark all resources
of that type as preparable. Additionally, TKO has built-in preparation for some GVK,
e.g. `topology.nephio.org`, `Placement`.

It is possible to avoid preparation for a specific preparable resource by setting
the `nephio.org/prepare` annotation. A common use case for this merges (see below).

Importantly, a preparation plugin works on the *entire* package, even though it is
triggered by a single resource. This means that the plugin can reference and even
update or delete other resources, as well as add new resources.

Not only that, but plugins get a TKO API client. This means that they can generate
new (child) deployments, templates, really anything that TKO can do. For example,
topology preparation generates child deployments based on templates and sites. Those
in turn can have additional topologies, which will be prepared in the next iteration,
allowing for any level of topology nesting.

Merges
------

A common requirement for generating new child deployments, e.g. for expanding
topologies, is that it's not enough to just copy over a template. We may also want
to introduce modifications or additions to that template. In fact, it may even be
*necessary* to do so if the template does not contain everything that is needed for
preparation and/or validation to complete successfully. We call this "merging".

Merging covers addition, injection, and updates. It is the way for a template to
accept "inputs" that it needs in order to be deployed successfully.

The prime example is the "merge" property of the topology `Placement` resource.
There we specify a list templates and for each provide a selector for its target
sites. The "merge" property per template lets us specific KRM in the topology
package (via ObjectReferences) that should be merged into that template.

Merging is straightforward but flexible. It starts by copying the template as is
into the deployment. The entire template can be modified during merging. New merged
resources will just be added. If a resource already exists (identified by GVK+name),
then it will be merged in: properties will be added or updated. This behavior can be
further controlled with the `nephio.org/merge` annotation, for example if you want to
replace the entire resource and not just merge properties. Additionally, the merged
name can be changed the `nephio.org/rename` annotation, letting you mix and match when
specifying multiple merges.

This scheme is powerful. It means that a topology package can include everything it
needs to create its child deployments. Not only that, when deploying a topology you
can further merge resources into it, so the topology itself can have "inputs".

Indeed, you can merge resources when creating a deployment via the API. This
includes the `--merge` flag for the `tko deployment create` CLI command.

This feature does introduce a small challenge. If you are including preparable
resources to merge in a package, how will the Preparer know not to prepare them
in the current deployment, and instead postpone preparation for the child
deployments? For this use case we go back to the `nephio.org/prepare` annotation.
When set to `Postpone` it means: don't prepare *here*, but prepare "later",
which in this case means in all the child deployments. (After merging, the
annotation value will be changed to `Here`.)

For an example, see the [O-RAN CU](examples/oran-cu/placement.yaml).
