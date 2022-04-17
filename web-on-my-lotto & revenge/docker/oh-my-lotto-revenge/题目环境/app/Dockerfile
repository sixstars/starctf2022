FROM python:3.8.12


WORKDIR /app

RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list \
    && sed -i 's/security.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list \
    && apt-get update

COPY source/requirements.txt .

RUN pip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple


ENV flag *ctf{ceaaa359cc578d2aaaf1f81db02d9964}
COPY source .
RUN mkdir guess

CMD gunicorn -c gunicorn.conf.py app:app
