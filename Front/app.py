from flask import Flask, request, jsonify, render_template, session
from flask_cors import CORS
from werkzeug.security import generate_password_hash, check_password_hash
import uuid

app = Flask(__name__)
app.secret_key = 'pesyn'
CORS(app)

users = {}
expressions = []

def current_user():
    return session.get('username')

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/api/v1/register', methods=['POST'])
def register():
    data = request.json
    username = data.get('username')
    password = data.get('password')
    if username in users:
        return jsonify({'error': 'User exists'}), 400
    users[username] = generate_password_hash(password)
    return jsonify({'message': 'Registered'})

@app.route('/api/v1/login', methods=['POST'])
def login():
    data = request.json
    username = data.get('username')
    password = data.get('password')
    pw_hash = users.get(username)
    if not pw_hash or not check_password_hash(pw_hash, password):
        return jsonify({'error': 'Invalid credentials'}), 401
    session['username'] = username
    return jsonify({'message': 'Logged in'})

@app.route('/api/v1/logout', methods=['POST'])
def logout():
    session.clear()
    return jsonify({'message': 'Logged out'})

@app.route('/api/v1/expressions', methods=['POST'])
def submit_expression():
    user = current_user()
    if not user:
        return jsonify({'error': 'Unauthorized'}), 401
    data = request.json
    expr = data.get('expression')
    expr_id = str(uuid.uuid4())
    record = {'id': expr_id, 'user': user, 'expr': expr, 'status': 'pending', 'result': None}
    expressions.append(record)
    # вычисление результата синхронно
    try:
        result = eval(expr, {'__builtins__': {}})
        record['result'] = result
        record['status'] = 'done'
    except Exception as e:
        record['result'] = str(e)
        record['status'] = 'error'
    return jsonify(record)

@app.route('/api/v1/expressions', methods=['GET'])
def get_expressions():
    user = current_user()
    if not user:
        return jsonify({'error': 'Unauthorized'}), 401
    user_exprs = [e for e in expressions if e['user'] == user]
    return jsonify(user_exprs)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)