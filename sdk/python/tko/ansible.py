import json, sys, os


tko_host = os.getenv('TKO_HOST', 'tko-data:50050')
max_count = int(os.getenv('MAX_COUNT', 1000))


class Inventory:
  def __init__(self):
    self.hosts = {}

  def add(self, name, host):
    self.hosts[name] = host

  def dump(self, out=sys.stdout):
    inventory = {'_meta': {'hostvars': self.hosts}}
    inventory['all'] = {'hosts': list(self.hosts.keys())}
    json.dump(inventory, out, separators=(',', ':'))
