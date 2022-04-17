from exts import db


class User(db.Model):

    __tablename__ = 'users'

    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(255), unique=True)
    password = db.Column(db.String(255), nullable=False)

class Note(db.Model):

    __tablename__ = 'notes'

    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(255), unique=False)
    note_id = db.Column(db.String(255), unique=True)
    text = db.Column(db.String(255), unique=False)
    title = db.Column(db.String(255), unique=False)

