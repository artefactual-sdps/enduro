# Documentation

These docs you are reading are built with MkDocs. This document describes the
configuration of the local environment and the general writing workflow.

## Environment configuration

Clone the repository:

    git clone https://github.com/artefactual-sdps/enduro

Access the documentation directory:

    cd enduro/docs

Create a Python virtual environment if it has not been created yet:

    python3 -m venv .venv

Enable the virtual environment:

    source .venv/bin/activate

Install the dependencies:

    pip install -r requirements.txt

Optionally, synchronize the environment:

    pip-sync

## Writing workflow

Run the builtin development server with live reloading support:

    mkdocs serve

The docs servers should be accessible under http://127.0.0.1:8000/.
