KRM Types in TKO
================

Note that these are *not* CRDs. They are types used for orchestration intent, intended to live
side-by-side with "real" Kubernetes resources and CRs in order to create a unified, declarative
experience. Technically they are never applied at edge clusters.

### metadata.nephio.org

This group is used to annotate templates and sites for specific filtering and matching.
Metadata is associated with these entities in the backend for efficient querying, e.g. in a table
in a SQL database that can be accessed with join queries.

The metadata keys will be gathered from the Kind and the nested spec keys with a "." path notation.

The name of the resource is *not* used for metadata, but is useful for identifying the resource for
merges. A good practice is to just name it "metadata".

Metadata gathering can be disabled per resource using the "nephio.org/metadata" annotation.

Note that metadata can also be added and modified directly against the backend. It is the *only*
out-of-band data supported (by definition metadata is out-of-band) and is minimalistic by design.

### deployment.nephio.org

Available Kind: Deployment, which will *always* be named "deployment".

This resource is automatically added to every deployment, if it's not already there, and is used for
information about the deployment, most notably to track the preparation status, to specify the template
that was used for it, the site it is associated with, etc.

### topology.nephio.org

This group is used for topology templates.

Placement: Used to create child deployments based on templates for sites. Sites can be specified via
the Site Kind, or selected via metadata filters, which allowing for targetting any number of sites.
Supports merging resources into the deployments. Note that child deployments are created as unprepared,
so each deployment will then be prepared for the particular target site. Also note that preparation can
happen already in the topology deployment, which is necessary for multi-sited preparation. Unprepared
resources will *not* be merged until they are prepared. The "nephio.org/prepare" and "nephio.org/merge"
annotations can be used to control this process.

Template: Used to select templates for Placement. Templates can be specified by name or looked up via
metadata filters.

Site: Used to select sites for Placement. Sites can be specified by name or looked up via metadata
filters. Additionally supported is provisioning *new* sites, which can be based on a site template.
Resources can be merged into newly provisioned sites similarly to how Placement merges resources
into deployments.

### workload.nephio.org

This group contains special workload types. The intent, as with all these types, is that they will
*not* be applied to edge clusters as CRs to be realized by a Kubernetes operator, but instead be
specially processed by instantiation plugins.

Available Kinds: HelmChart

### infra.nephio.org

This group is used to specify infrastructure requirements. Infra resources can appear in all entity
types: templates, sites, and deployments.

During instantiation, the expectation is that the instantiation plugin will take into account all
infra requirements for the site and its deployments.

### [name].plugin.nephio.org

This group family is used to configure plugins, for both deployment preparation and site instantiation.
Individual GVKs can be used as targets for the relevant plugin.

Note that plugins can target *any* GVK and do not need to make use of this group family, and that
beyond the target resource the plugin can make use of *any* resource in the deployment.

The [name] can be product- or vendor-specific, e.g. "free5gc" or "nokia".

Annotations
===========

| Annotation           | Values                          | Description |
|----------------------|---------------------------------|-------------|
| `nephio.org/metadata`  | `Here` (default), `Postpone`, `Never` | `Never` disables use of a metadata resource. `Postpone` will disable use here, and reenable after merging elsewhere. |
| `nephio.org/merge`     | `Replace` (default), `Override`     | `Override` will override individual properties instead of replacing the whole resource. |
| `nephio.org/prepared`  | `false` (default), `true`           | `true` marks a resource as prepared. |
| `nephio.org/prepare`   | `Here` (default), `Postpone`, `Never` | `Never` disables preparation of a target resource. `Postpone` will disable preparation here, and reenable after merging elsewhere. |
| `nephio.org/rename`    | | Will give the resource a new name when merged. |

Merges
======

Both Placement and Site support a `merge` keyname, which is a list of
[ObjectReferences](https://dev-k8sref-io.web.app/docs/common-definitions/objectreference-/).

During preparation, *all* resources to be merged must also be prepared first, otherwise preparation
will be aborted. If a resource to be merged has `Postpone` as the value of the `nephio.org/metadata`
or `nephio.org/prepare` annotations then that value will be changed to `Here` after the merge.

The resource can be renamed during merge using the `nephio.org/rename` annotation.

The merging semantics can be controlled via the `nephio.org/merge` annotation.
