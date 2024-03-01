ECHO OFF
ECHO Setting up environment...
IF NOT EXIST .\.venv python -m venv .\.venv
CALL .\.venv\Scripts\activate
CALL .\.venv\Scripts\pip install -r requirements.txt
