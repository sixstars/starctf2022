workers = 10 
worker_class = "gevent"
bind = "0.0.0.0:6680"
capture_output = True

accesslog = '/home/ubuntu/oh-my-lotto/log/gunicorn.access.log'
access_log_format = '%(h)s %(l)s %(u)s %(t)s'

errorlog = '/home/ubuntu/oh-my-lotto/log/gunicorn.error.log'
loglevel = 'debug'