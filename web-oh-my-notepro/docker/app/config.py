from pymysql.constants import CLIENT

class Config(object):
    SECRET_KEY = 'you-will-never-guess-hahahahafeffefefefefefxwdhaha2333'
    # SQLALCHEMY_DATABASE_URI = 'mysql+pymysql://root:root@mysql:3306/ctf?charset=utf8mb4&local_infile=1'
    SQLALCHEMY_DATABASE_URI = 'mysql+pymysql://ctf3:ctf123456@mysql:3306/ctf?charset=utf8mb4&local_infile=1'
    SQLALCHEMY_ENGINE_OPTIONS = {"connect_args":{"client_flag": CLIENT.MULTI_STATEMENTS}}
    SQLALCHEMY_POOL_RECYCLE = 30
    SQLALCHEMY_POOL_SIZE = 40