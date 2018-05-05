import requests
import pymongo
url = 'https://127.0.0.1:8080'
subnets = [
    ('china office', '172.16.9.0', 24),
    ('russia office', '172.16.10.0', 24),
    ('LA office', '172.16.11.0', 24),
    ('Prod', '10.100.0.0', 16),
    ('Prod LA', '10.100.11.0', 24)
]

devices = [
      ('HTTP', 'OPS', 'ksam', 'china', [('eth0', 'AA:BB:CC:DD:EE:FF', [('172.16.9.1', 1)])]),
      ('DNS', 'OPS', 'ksam', 'china', [('eth0', 'AA:BB:CC:DD:EE:EE', [('172.16.9.53', 1)])]),
      ('SQL', 'OPS', 'jnottingham', 'russia', [('eth0', 'AA:BB:CC:DD:EE:DD', [('172.16.10.1', 2)])]),
      ('Router', 'Networking', 'jgem', 'LA', [('eth0', 'AA:BB:CC:DD:EE:CC', [('172.16.11.124', 3)]),      ('eth1', 'BB:BB:BB:BB:BB:BB', [('10.100.100.254', 4)])]),
      ('AD', 'SysAd', 'bwhite', 'laptop', [('eth0', 'AA:BB:CC:DD:EE:BB', [('172.16.10.101', 2)])]),
      ('Jenkins oh god', 'Engineering', 'rgeorge', 'LA', [('eth0', 'AA:BB:CC:DD:EE:AA', [('10.100.100.200', 4)])]),
      ('gitlab', 'Engineering', 'rgeorge', 'LA', [('eth0', 'AA:BB:CC:DD:FF:FF', [('10.100.100.201',       4)])]),
      ('esxi', 'OPS', 'ksam', 'china', [('eth0', 'AA:BB:CC:DD:FF:EE', [('10.100.100.202', 4)])])
]

new_subnets = [
    ('Prod TEST', '10.100.100.0', 24)
]

def add_subnets():
    for subnet in subnets:
        sub = {
                 'subnet_name': subnet[0],
                 'ip': subnet[1],
                 'mask': subnet[2],
                 'cidr': '{}/{}'.format(subnet[1], subnet[2])
             }
        requests.put('{}/api/v1/subnets'.format(url), json=sub, verify=False)

def add_devices():
    for device in devices:
        interfaces = []
        for interface in device[4]:
            interfaces.append({
                'interface_name': interface[0],
                'mac': interface[1],
                'ips': [{
                    'ip': interface[2][0][0]
                }]
            })
        dev = {
            'device_name': device[0],
            'team': device[1],
            'owner': device[2],
            'location': device[3],
            'interfaces': interfaces
        }
        requests.put('{}/api/v1/devices'.format(url), json=dev, verify=False)

def add_new_subnets():
    for subnet in new_subnets:
        sub = {
                 'subnet_name': subnet[0],
                 'ip': subnet[1],
                 'mask': subnet[2],
                 'cidr': '{}/{}'.format(subnet[1], subnet[2])
             }
        requests.put('{}/api/v1/subnets'.format(url), json=sub, verify=False)

def main():
    client = pymongo.MongoClient()
    client.drop_database("changan_test")
    add_subnets()
    add_devices()
    add_new_subnets()

if __name__ == '__main__':
    main()
