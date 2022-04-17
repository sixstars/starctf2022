FROM ubuntu:20.04
RUN sed -i "s/http:\/\/archive.ubuntu.com/http:\/\/mirrors.ustc.edu.cn/g" /etc/apt/sources.list
RUN apt-get update && apt-get -y upgrade
RUN apt-get install -y lib32z1 xinetd
copy musl_1.2.2-1_amd64.deb /musl_1.2.2-1_amd64.deb
RUN dpkg -i /musl_1.2.2-1_amd64.deb
RUN useradd -u 8888 -m pwn
CMD ["/usr/sbin/xinetd", "-dontfork"]
