import requests
import populate_test_db
import populate_test_mongo
import logging
from IPython import embed
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

url_base = "https://127.0.0.1:8080/api/v1/"
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def equal_dicts(d1, d2, ignore_keys):
    ignored = set(ignore_keys)
    for k1, v1 in d1.iteritems():
        if k1 not in ignored and (k1 not in d2 or d2[k1] != v1):
            return False
    for k2, v2 in d2.iteritems():
        if k2 not in ignored and k2 not in d1:
            return False
    return True

def test_devices():
    url = '{}devices'.format(url_base)
    # Test get
    r = requests.get(url, verify=False)
    if r.status_code != 200:
        logger.error("devices get did not succeed")
        return
    base_devices = r.json()['devices']
    if len(base_devices) == 0:
        logger.error("devices is empty")
    logger.debug("base devices:")
    for device in base_devices:
        logger.debug("device: {}".format(device))

    # Test add
    device = {
            "device_name": "test_device",
            "team": "test_team",
            "owner": "test_owner",
            "location": "test_location",
            "interfaces": [{
                "mac": "ff:ff:ff:ff:ff:ff",
                "interface_name": "eth0",
                "ips": [{
                    "ip":"8.8.8.8"
                    }]
            }]
    }
    r = requests.put(url, json=device, verify=False)
    logger.debug(r.text)
    r = requests.get(url, verify=False)
    devices = r.json()['devices']
    if not any(equal_dicts(d, device, ['device_id', 'interfaces']) for d in devices):
        logger.error("devices put/add failed")

    # Test delete
    r = requests.delete(url, json=device, verify=False)
    r = requests.get(url, verify=False)
    devices = r.json()['devices']
    if any(d['device_name'] == 'test_device' for d in devices):
        logger.error("devices delete failed")

    # Test edit
    device = base_devices[1]
    device['device_name'] = "a_test_device"
    r = requests.post(url, json=device, verify=False)
    logger.debug(r.json())
    r = requests.get(url, verify=False)
    devices = r.json()['devices']
    if not any(d['device_name'] == 'a_test_device' for d in devices):
        logger.error("devices post/edit failed")

def test_subnets():
    url = '{}subnets'.format(url_base)

    # Test get
    r = requests.get(url, verify=False)
    if r.status_code != 200:
        logger.error("subnets get did not succeed")
        return
    base_subnets = r.json()['subnets']
    if len(base_subnets) == 0:
        logger.error("subnets is empty")

    # Test add
    subnet = {
            "subnet_name": "test_subnet",
            "ip": "8.8.8.8",
            "mask": 32
            }
    r1 = requests.put(url, json=subnet, verify=False)
    r = requests.get(url, verify=False)
    subnets = r.json()['subnets']
    if not any(equal_dicts(s, subnet, ['subnet_id']) for s in subnets):
        logger.error("subnets put/add failed")
        logger.error("{}".format(r1.json()))

    # Test delete
    r = requests.delete(url, json=subnet, verify=False)
    r = requests.get(url, verify=False)
    subnets = r.json()['subnets']
    if any(s['subnet_name'] == 'test_subnet' for s in subnets):
        logger.error("subnets delete failed")

    # Test edit
    subnet = base_subnets[1]
    subnet['subnet_name'] = "test_subnet"
    r = requests.post(url, json=subnet, verify=False)
    r = requests.get(url, verify=False)
    subnets = r.json()['subnets']
    if not any(s['subnet_name'] == 'test_subnet' for s in subnets):
        logger.error("subnets post/edit failed")

def test_reports():
    url = '{}reports'.format(url_base)

    report = {
            "title": "test_report",
            "report": "The full report as a test",
            "ip": "172.16.9.53"
            }
    # Test put
    r = requests.put(url, json=report, verify=False)
'''
def test_ips():
    url = '{}ips'.format(url_base)

    # Test get
    r = requests.get(url)
    if r.status_code != 200:
        logger.error("ips get did not succeed")
        return
    base_ips = r.json()['ips']
    if len(base_ips) == 0:
        logger.error("ips is empty")

    # Test add
    ip = {
            "name": "test_ip",
            "ip": "8.8.8.8"
            }
    r = requests.put(url, json=ip)
    r = requests.get(url)
    ips = r.json()['ips']
    if not any(i['name'] == 'test_ip' and i['ip'] == '8.8.8.8' for i in ips):
        logger.error("ips put failed")

    # Test delete
    r = requests.delete(url, json=ip)
    r = requests.get(url)
    ips = r.json()['ips']
    if any(i['name'] == 'test_ip' and i['ip'] == '8.8.8.8' for i in ips):
        logger.error("ips delete failed")

    # Test edit
    ip = base_ips[1]
    ip['name'] = "test_ip"
    r = requests.post(url, json=ip)
    r = requests.get(url)
    ips = r.json()['ips']
    if not any(i['name'] == 'test_ip' for i in ips):
        logger.error("ips post failed")
'''

def main():
    #populate_test_db.main()
    populate_test_mongo.main()
    test_devices()
    test_subnets()
    test_reports()
    #test_ips()

if __name__ == '__main__':
    main()
