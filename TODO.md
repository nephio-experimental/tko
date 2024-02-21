TODO
====

* Implement tko-runner
* Archive previous versions of templates, sites, and deployments to allow for reverting
  changes
* Instantiation: keep track of deleted deployments per site
* If Placement uses a site selector, should we continuously track the selection? How?
* Parallel processing by multiple controller instances (sharding?)
* Run plugins safely on 1) remote machines, 2) containers, 3) both
* Paging of backend results (gRPC does streaming, but can we also stream the backend API?
  How will that work with huge SQL query results?)
* Support Spanner
* Support TOSCA ETSI NSDs
