WARNING: This repository is published by the [Nephio Authors](https://nephio.org/) but is
neither endorsed nor maintained by the Nephio Technical Steering Committee (TSC). It is intended
to be used for reference only. The Nephio distribution repositories are located in the
[`nephio-project` organization](https://github.com/nephio-project). For more information
[see this page](https://nephio.org/experimental).

TKO
===

A PoC demonstrating scalability options for Nephio with a focus on decoupling the various
subsystems, especially the data backend and API access.

* [Nephio R2 Summit Presentation](https://docs.google.com/presentation/d/1I54I6RvexMjcP-qJSDq3xEqCyD6rCEfU_lcAdSFy1iM)
* [Installation guide](INSTALL.md)
* [KRM types](KRM.md)
* [TODO](TODO.md)

Usage Example
-------------

To access the web UI go to [http://localhost:50051](http://localhost:50051).

The `tko` CLI is quite rich. Very quick example:

    tko deployment list --template-metadata=NetworkFunction.family=5G*

To see the service logs (you can do this in separate terminal tabs):

    scripts/log-service tko-api-server --follow
    scripts/log-service tko-preparer --follow
    scripts/log-service tko-meta-scheduler --follow

To see the instantiated KIND clusters:

    kind get clusters

To access a KIND cluster use a context name where the cluster name is prefixed with "kind-", e.g.:

    kubectl get pods --all-namespaces --context=kind-edge1
