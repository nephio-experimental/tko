#!/usr/bin/env python3

import tko, tko.ansible


inventory = tko.ansible.Inventory()

with tko.Client(host=tko.ansible.DATA_HOST) as client:
  for listed_site in client.list_sites(offset=0, max_count=tko.ansible.MAX_COUNT):
    site = client.get_site(listed_site.siteId)
    inventory.add(site.siteId, {
      'template_id': site.templateId,
      'metadata': dict(site.metadata),
      'updated': site.updated.ToJsonString(),
      'deployment_ids': list(site.deploymentIds),
      'package': site.get_package()
    })

inventory.dump()
