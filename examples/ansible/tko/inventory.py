#!/usr/bin/env python3

import json, sys


class Hosts:
  def __init__(self):
    self.hosts = {}

  def add(self, name, host):
    self.hosts[name] = host

  def dump(self):
    inventory = {'_meta': {'hostvars': self.hosts}}
    inventory['all'] = {'hosts': list(self.hosts.keys())}
    json.dump(inventory, sys.stdout, separators=(',', ':'))
  

hosts = Hosts()
hosts.add('web1.example.com', {'ansible_host': '1.2.3.4', 'testvar': 'hi'})
hosts.add('web2.example.com', {'ansible_host': '4.3.2.1', 'testvar': 'ho'})
hosts.dump()
