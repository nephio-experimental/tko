WARNING: This repository is published by the [Nephio Authors](https://nephio.org/) but is
neither endorsed nor maintained by the Nephio Technical Steering Committee (TSC). It is intended
to be used for reference only. The Nephio distribution repositories are located in the
[`nephio-project` organization](https://github.com/nephio-project). For more information
[see this page](https://nephio.org/experimental).

TKO
===

A PoC demonstrating scalability opportunities for Nephio with a focus on decoupling the various
subsystems, specifically the data backend and API access, as well as integration with external
site inventories and blueprint catalogs.

Included backends are for [PostgreSQL](https://www.postgresql.org/) and
[Spanner](https://cloud.google.com/spanner). Both provide scalability, resiliency, and
atomic updates via transactions. (Note: Spanner backend is work-in-progress.) It is entirely
possible to create a backend based on git (a.k.a. "GitOps"). Such an implementation may be
suitable for storing templates, but is probably not a good idea for sites and deployments,
which are expected to number in the millions in production environments.

Access to the API is via [gRPC](https://grpc.io/), which is widely supported, including in
loadbalancers (see
[Envoy](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_protocols/grpc))
and thus service meshes. TKO makes good use of gRPC streaming directly from the
backend to the clients, allowing scalable access to extremely large result sets.

This PoC is a complete rewrite of the Nephio core. It comprises three controllers that can run
independently or be embedded in a control plane, such as a Kubernetes management cluster.
Note that when running in Kubernetes TKO does *not* use Kubernetes's API or its etcd data
store. However, it is possible to implement a meta-scheduling plugin that would create
resources on the management cluster, e.g. to interact with a Kubernetes-native infrastructure
manager.

Other ways in which TKO differs from Nephio:

* A different approach to "specialization" (here called "preparation"), replacing the
  [kpt](https://kpt.dev/) file with per-resource plugins: no pipeline, no conditions.
  Enforces atomic updates. Kpt functions still get first-class support.
* A different approach to "instantiation" based on meta-scheduling of complete sites
  together with *all* their associated workload deployments, including infrastructure
  resources for infrastructure managers. Can work with existing sites and also provision
  new ones.
* A different approach to topologies, incorporating topology decomposition as part of the
  preparation process. Topologies can be nested. Also supports
  [TOSCA](https://www.oasis-open.org/committees/tosca/) topologies as an alternative frontend,
  implemented with [Puccini](https://github.com/tliron/puccini).
* A different approach to KRM validation, supporting custom validation plugins that can
  go far beyond OpenAPIv3 schemas. Can also work on templates via "partial" validation. Invalid
  KRM is *never* allowed into the backend. Relies on
  [Kubeconform](https://github.com/yannh/kubeconform) for validating standard Kubernetes
  resources via their published OpenAPIv3 schemas.

Additional features demonstrated:

* SDK for Python-based plugins for preparation, meta-scheduling, and validation. The SDK
  does most of heavy lifting, allowing devs to focus on the network function or cloud
  platform vendor logic.
* Rich metadata support, enabling powerful and scalable package querying and template/site
  selection for topologies.
* Support for [Helm](https://helm.sh/) charts with an expression language for pulling in
  chart input values (as an alternative to full-blown preparation plugins).
* Web dashboard.
* Rich terminal dashboard, with mouse support.
* IPv6 first, with support for dual-stack IP.
* Unified structured logging via [CommonLog](https://github.com/tliron/commonlog),
  including support for logging to journald.

Documentation
-------------

* [Installation guide](INSTALL.md)
* [User guide](USAGE.md)
* [Reference guide](REFERENCE.md)
* [How preparation works](PREPARATION.md)
* [TODO](TODO.md)

Architecture Diagram
--------------------

```mermaid
%%{init: {'themeVariables': { 'edgeLabelBackground': 'transparent'}}}%%
flowchart TD
    D[(Data Backends)]

    C(TKO CLI and TUI)
    W(TKO Web GUI)
    O(Orchestrators)

    A[TKO API Server]
    P[TKO Preparer]
    M[TKO Meta-Scheduler]

    PP[/prepare plugins/]
    VP[/validate plugins/]
    SP[/schedule plugins/]

    C-. gRPC ..-A
    W-. HTTP ..-A
    O-. gRPC ..-A
    A-. gRPC .- P
    A-. gRPC .-M

    D-..-A

    PP---P
    VP---P
    SP---M

    P-.-DPA
    M-.-DPA
    M-.-WC

    subgraph TC [Template Catalog]
        ST[\Site Templates/]
        WT[\Workload Templates/]
        TT[\Topology Templates/]
    end

    subgraph SI [Site Inventory]
        PS{{Provisioned Sites}}
        ES{{Existing Sites}}
    end

    subgraph DPA [Deployment Preparation Area]
        UT{{Unprepared Topologies}}
        UR{{Unprepared Resources}}
        PR{{Prepared Resources}}
    end

    subgraph WC [Workload Cluster]
        SR{{Scheduled Resources}}
    end

    TT-->UT-->UR-->PR-->SR
    WT-->UR
    ST-->PS-->WC

    style A fill:blue,color:white
    style C fill:blue,color:white
    style W fill:blue,color:white
    style P fill:blue,color:white
    style PP fill:darkblue,color:white
    style VP fill:darkblue,color:white
    style M fill:blue,color:white
    style SP fill:darkblue,color:white
    style ST fill:red,color:white
    style TT fill:red,color:white
    style WT fill:red,color:white
    style UT fill:yellow
    style UR fill:yellow
    style PR fill:green,color:white
    style SR fill:green,color:white
    style PS fill:green,color:white
    style ES fill:green,color:white
    linkStyle 0,1,2,3,4,9,10 stroke:blue,color:blue
    linkStyle 6,7,8 stroke:darkblue,color:darkblue
```


Additional Resources
--------------------

* [Components proposal](https://docs.google.com/drawings/d/1I7e3zm9-xC6cDxNd_ANPCVGbOQLgjAX25cImhwObG74)
  ([presentation recording](https://www.youtube.com/watch?v=nwd4t0DTTH8))
* [Nephio R2 Summit presentation slides](https://docs.google.com/presentation/d/1I54I6RvexMjcP-qJSDq3xEqCyD6rCEfU_lcAdSFy1iM)
