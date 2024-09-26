#!/usr/bin/env python3

import tko, tko.ansible, os


metadata_patterns = os.getenv('TKO_METADATA')
if metadata_patterns:
  metadata_patterns = tko.yaml.load(metadata_patterns)

inventory = tko.ansible.Inventory()

with tko.Client(host=tko.ansible.DATA_HOST) as client:
  for listed_site in client.list_sites(offset=0, max_count=tko.ansible.MAX_COUNT, metadata_patterns=metadata_patterns):
    site = client.get_site(listed_site.siteId)
    inventory.add(site.siteId, {
      'template_id': site.templateId,
      'metadata': dict(site.metadata),
      'updated': site.updated.ToJsonString(),
      'deployment_ids': list(site.deploymentIds),
      'package': site.get_package()
    })

inventory.dump()
