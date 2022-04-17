from flask_wtf import FlaskForm
from wtforms import *
from wtforms.validators import DataRequired


class CreateNoteForm(FlaskForm):
    title = StringField('Note Title', validators=[DataRequired()])
    body = TextAreaField('Write something', validators=[DataRequired()])
    submit = SubmitField('Post!')

class CreateLoginForm(FlaskForm):
    username = StringField('Enter username', validators=[DataRequired()])
    password = StringField('Enter password', validators=[DataRequired()])
    submit = SubmitField('Login!')


