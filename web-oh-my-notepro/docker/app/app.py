import string
import random
from flask import render_template, redirect, url_for, request, session, Flask
from functools import wraps

from exts import db
from config import Config
from models import User, Note
from forms import CreateNoteForm, CreateLoginForm
from utils import md5

app = Flask(__name__)
app.config.from_object(Config)
app.config['MYSQL_LOCAL_INFILE'] = True
db.init_app(app)


def login_required(f):
    @wraps(f)
    def decorated_function(*args, **kws):
            if not session.get("username"):
               return redirect(url_for('login'))
            return f(*args, **kws)
    return decorated_function


def get_random_id():
    alphabet = list(string.ascii_lowercase + string.digits)
    return ''.join([random.choice(alphabet) for _ in range(32)])

@app.route('/')
@app.route('/index')
@login_required
def index():
    username = session['username']
    results = Note.query.filter_by(username=username).limit(100).all()
 
    notes = []
    for x in results:
        note = {}
        note['title'] = x.title
        note['note_id'] = x.note_id
        notes.append(note)

    return render_template('index.html', notes=notes)


@app.route('/login', methods=["GET", "POST"])
def login():
    form = CreateLoginForm()
    if session.get('username'):
        return redirect(url_for('index'))
    if request.method == 'GET':
        return render_template('login.html', form=form)
    else:
        username = form.username.data
        password = form.password.data
        
        password_md5 = md5(password)

        user = User.query.filter_by(username=username).first()

        if user:
            if password_md5 != user.password:
                return render_template('login.html', form=form, message='Sorry, username or password ERROR!')
            else:
                session['username'] = username
                return redirect(url_for('index'))
        else:
            user = User(username=username, password=password_md5)
            db.session.add(user)
            db.session.commit()
            session['username'] = username

        return redirect(url_for('index'))

@app.route('/logout')
@login_required
def logout():
    session.pop('username', None)
    return redirect(url_for('login'))

@app.route('/create', methods=['GET', 'POST'])
@login_required
def create():
    try:
        form = CreateNoteForm()
        if request.method == "POST":
            username = session['username']
            title = form.title.data
            text = form.body.data
            note_id = get_random_id()
            note = Note(username=username,
                        title=title, text=text,
                        note_id=note_id)
            db.session.add(note)
            db.session.commit()
            return redirect(url_for('index')) 
        else:
            return render_template("create.html", form=form) 
    except Exception as e:
        return str(e)


@app.route('/view')
@login_required
def view():
    note_id = request.args.get("note_id")
    sql = f"select * from notes where note_id='{note_id}'" 
    print(sql)
    result = db.session.execute(sql, params={"multi":True})
    db.session.commit()
    
    result = result.fetchone()
    data = {
        'title': result[4],
        'text': result[3],
    }
    return render_template('note.html', data=data)


if __name__ == '__main__':
    app.run(threaded=False, processes=100, debug=True, host='0.0.0.0', port=5000)