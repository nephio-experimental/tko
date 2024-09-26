import json, sys, os, tko.encoding


DATA_HOST = os.getenv('TKO_DATA_HOST', 'tko-data:50050')
MAX_COUNT = int(os.getenv('TKO_MAX_COUNT', 100))
METADATA_PATTERNS = os.getenv('TKO_METADATA_PATTERNS')

if METADATA_PATTERNS:
  METADATA_PATTERNS = tko.encoding.yaml.load(METADATA_PATTERNS)


class Inventory:
  def __init__(self):
    self.hosts = {}

  def add(self, name, host):
    self.hosts[name] = host

  # The format expected by Ansible inventory plugins
  def dump(self, out=sys.stdout):
    inventory = {'_meta': {'hostvars': self.hosts}}
    inventory['all'] = {'hosts': list(self.hosts.keys())}
    json.dump(inventory, out, separators=(',', ':'))
