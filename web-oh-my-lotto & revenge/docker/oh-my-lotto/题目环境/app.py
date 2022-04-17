
from flask import Flask, render_template, request, session
import subprocess
import time
import random
import socket
import hashlib

s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
s.connect(("8.8.8.8", 80))
ip = s.getsockname()[0]

def md5(data):
    m = hashlib.md5(data.encode())
    return m.hexdigest()


app = Flask(__name__)
app.config['SECRET_KEY'] = "SECRET_KEYxefexsxfwefvwevwvwev"
time_list = []
in_use = []

@app.route("/", methods=['GET', 'POST'])
def index():
    
    if request.method == 'GET':
        n = str(random.randint(2000000, 8000000))
        session['n'] = n
        print(n)
        n_md5 = md5(n)
        session['n_md5'] = n_md5[:6]
        return render_template('index.html', md5=session['n_md5'])
        
    elif request.method == 'POST':

        n = request.form.get('n')
        print(n)
        a = md5(n)
        b = md5(session['n'])
        a = a[:6]
        b = b[:6]
        if a != b:
            return render_template('index.html', md5=session['n_md5'], message='incorrect n')
        else:
            
            if len(time_list) == 0:
                for i in range(0, 60):
                    time_list.append(time.time())
                    in_use.append(False)

            index = -1
            for i in range(0, 60):
                now_time = time.time()
                last_time = time_list[i]
                if now_time - last_time > 180 and in_use[i] == True:
                    time_list[i] = now_time
                    in_use[i] = False
                    continue
                if in_use[i] == False:
                    time_list[i] = now_time
                    in_use[i] = True
                    index = i
                    break
            if index != -1:
                subprocess.Popen("./run.sh " + str(index), shell=True)
        
            port = str(53000 + index)
            print("now index is: ", index)
            return render_template('index.html', md5=session['n_md5'], ip=ip, port=port)
    

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=6690)
