# README

* docker - source code

* oh-my-grafana-EN.md(English Writeup)

* oh-my-grafana-ZH.md(Chinese Writeup)

oh-my-grafana这个题预期是已经把改密码的功能ban掉了，所以师傅们在做的时候应该不是有人在故意改密码，是因为grafana有个`disable_brute_force_login_protection`的配置项，默认如果尝试登录某账号密码错误5次，该账号将会被锁定，后期把这个配置项关闭以后应该做题就正常了，给师傅们做题添麻烦了。

oh-my-grafana is expected to have ban the function  of changing the password, so when the you are doing it, it should not be that someone is changing the password, because grafana has a `disable_brute_force_login_protection` configuration item. By default, if you try to log in to an account with the wrong password for 5 times, the account will be locked. After closing this configuration item later, it should be normal to do this challenge, sorry for that.