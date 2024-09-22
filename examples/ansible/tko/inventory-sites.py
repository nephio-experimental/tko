#!/usr/bin/env python3

import tko, tko.ansible


inventory = tko.ansible.Inventory()

with tko.Client(host=tko.ansible.tko_host) as client:
  for listed_site in client.list_sites(offset=0, max_count=tko.ansible.max_count):
    inventory.add(listed_site.siteId, {
      'template_id': listed_site.templateId,
      'metadata': dict(listed_site.metadata),
      'updated': listed_site.updated.ToJsonString(),
      'deployment_ids': list(listed_site.deploymentIds)
    })

inventory.dump()
