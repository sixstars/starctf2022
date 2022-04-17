import hashlib

def md5(data):
    m = hashlib.md5(data.encode())
    return m.hexdigest()