#!/usr/bin/env python3

import tko, tko.ansible


inventory = tko.ansible.Inventory()

with tko.Client(host=tko.ansible.DATA_HOST) as client:
  for listed_site in client.list_sites(offset=0, max_count=tko.ansible.MAX_COUNT, metadata_patterns=tko.ansible.METADATA_PATTERNS):
    site = client.get_site(listed_site.siteId).to_ard()
    site_id = site.pop('site_id')
    inventory.add(site_id, site)

inventory.dump()
