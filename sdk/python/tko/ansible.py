import json, sys, os


DATA_HOST = os.getenv('TKO_DATA_HOST', 'tko-data:50050')
MAX_COUNT = int(os.getenv('TKO_MAX_COUNT', 100))


class Inventory:
  def __init__(self):
    self.hosts = {}

  def add(self, name, host):
    self.hosts[name] = host

  def dump(self, out=sys.stdout):
    inventory = {'_meta': {'hostvars': self.hosts}}
    inventory['all'] = {'hosts': list(self.hosts.keys())}
    json.dump(inventory, out, separators=(',', ':'))
