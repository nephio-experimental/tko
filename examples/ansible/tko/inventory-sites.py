#!/usr/bin/env python3

import tko, tko.ansible, os


tko_host = os.getenv('TKO_HOST', 'tko-data:50050')
max_count = int(os.getenv('MAX_COUNT', 1000))

inventory = tko.ansible.Inventory()

with tko.Client(host=tko_host) as client:
  for site in client.list_sites(offset=0, max_count=max_count):
    inventory.add(site.siteId, {
      'template_id': site.templateId,
      'metadata': dict(site.metadata),
      'updated': site.updated.ToJsonString(),
      'deployment_ids': list(site.deploymentIds)
    })

inventory.dump()
