WARNING: This repository is published by the [Nephio Authors](https://nephio.org/) but is
neither endorsed nor maintained by the Nephio Technical Steering Committee (TSC). It is intended
to be used for reference only. The Nephio distribution repositories are located in the
[`nephio-project` organization](https://github.com/nephio-project). For more information
[see this page](https://nephio.org/experimental).

TKO
===

A PoC demonstrating scalability options for Nephio with a focus on decoupling the various
subsystems, specifically the data backend and API access, as well as integration with external
site inventories and blueprint catalogs.

Additional features demonstrated:

* A different approach to "specialization" (here called "preparation"), replacing the kpt
  pipeline with per-resource plugins: no pipeline, no conditions. Enforces atomic updates.
* A different approach to "instantiation" based on meta-scheduling for a complete site
  together with *all* its associated workload deployments, including infrastructure KRMs
  for infrastructure managers. Can work with existing sites and also provision new ones.
* A different approach to topologies, incorporating topology decomposition as part of the
  preparation process. Also supports TOSCA topologies as an alternative frontend,
  via [Puccini](https://github.com/tliron/puccini).
* A different approach to KRM validation, supporting custom validation plugins that go way
  beyond OpenAPIv3. Can also work on templates via "partial" validation. Invalid KRM is
  *never* allowed into the backend. Relies on
  [Kubeconform](https://github.com/yannh/kubeconform) for standard Kubernetes resources.
* SDK for Python-based plugins for preparation, meta-scheduling, and validation. The SDK
  does most of heavy lifting, allowing devs to focus on the network function or cloud platform
  vendor logic.
* Rich metadata support, enabling for powerful and scalable package querying and
  template/site selection.
* IPv6 first, with powerful support for dual-stack IP.
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

    C(TKO CLI)
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
