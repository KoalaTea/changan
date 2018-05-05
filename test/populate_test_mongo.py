#! /usr/bin/python
import time
from pymongo import MongoClient, ASCENDING
from werkzeug.security import generate_password_hash
# TODO make usernames unique

# subnet_id subject to change
device_objs = [
     #(name, team, owner, location, interfaces[(name, mac, ips[ip])])
     ('HTTP', 'OPS', 'ksam', 'china', [('eth0', 'AA:BB:CC:DD:EE:FF', [('172.16.9.1', 1)])]),
     ('DNS', 'OPS', 'ksam', 'china', [('eth0', 'AA:BB:CC:DD:EE:EE', [('172.16.9.53', 1)])]),
     ('SQL', 'OPS', 'jnottingham', 'russia', [('eth0', 'AA:BB:CC:DD:EE:DD', [('172.16.10.1', 2)])]),
     ('Router', 'Networking', 'jgem', 'LA', [('eth0', 'AA:BB:CC:DD:EE:CC', [('172.16.11.124', 3)]), ('eth1', 'BB:BB:BB:BB:BB:BB', [('10.100.100.254', 4)])]),
     ('AD', 'SysAd', 'bwhite', 'laptop', [('eth0', 'AA:BB:CC:DD:EE:BB', [('172.16.10.101', 2)])]),
     ('Jenkins oh god', 'Engineering', 'rgeorge', 'LA', [('eth0', 'AA:BB:CC:DD:EE:AA', [('10.100.100.200', 4)])]),
     ('gitlab', 'Engineering', 'rgeorge', 'LA', [('eth0', 'AA:BB:CC:DD:FF:FF', [('10.100.100.201', 4)])]),
     ('esxi', 'OPS', 'ksam', 'china', [('eth0', 'AA:BB:CC:DD:FF:EE', [('10.100.100.202', 4)])])
]

subnet_ids = {}

'''
subnet_objs = [
    ('china office', '172.16.9.0', 24),
    ('russia office', '172.16.10.0', 24),
    ('LA office', '172.16.11.0', 24),
    ('Prod', '10.100.0.0', 16)
]
'''
subnet_objs = [] #moves to a dynamic adder

def insert_data(db):
    subnets = db.subnets
    for sub in subnet_objs:
        subnets.insert_one(
                {
                    'subnet_name': sub[0],
                    'ip': sub[1],
                    'mask': sub[2]
                }
        )
    subnets_objs = subnets.find()
    for obj in subnets_objs:
        if obj['subnet_name'] == 'china office':
            subnet_ids[1] = obj['_id']
        elif obj['subnet_name'] == 'russia office':
            subnet_ids[2] = obj['_id']
        elif obj['subnet_name'] == 'LA office':
            subnet_ids[3] = obj['_id']
        elif obj['subnet_name'] == 'Prod':
            subnet_ids[4] = obj['_id']
    devices = db.devices
    for dev in device_objs:
        interfaces = []
        for interface in dev[4]:
            interfaces.append({
                'interface_name': interface[0],
                'mac': interface[1],
                'ips': [{
                    'ip': interface[2][0][0],
                    'subnet_id': subnet_ids[interface[2][0][1]]
                }]
            })
        devices.insert_one(
                {
                    'device_name': dev[0],
                    'team': dev[1],
                    'owner': dev[2],
                    'location': dev[3],
                    'interfaces': interfaces
                }
        )

def main():
    client = MongoClient()
    client.drop_database("changan_test")
    db = client.changan_test

    #insert_data(db)

if __name__ == '__main__':
    main()
