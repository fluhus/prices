# Runs pep8 and pylint on source files, with project-specific settings.
# PLEASE RUN THIS BEFORE COMMITTING CODE.
pep8 --ignore=W391 . &&
pylint -d invalid-name,trailing-newlines,fixme *.py

