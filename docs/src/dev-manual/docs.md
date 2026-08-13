# Documentation

These docs you are reading are built with MkDocs. This document describes the
configuration of the local environment and the general writing workflow.

## Environment configuration

Clone the repository:

    git clone https://github.com/artefactual-sdps/enduro

Access the documentation directory:

    cd enduro/docs

Install [uv](https://docs.astral.sh/uv/getting-started/installation/) if it is
not already available, then synchronize the project environment:

    uv sync

## Writing workflow

Run the builtin development server with live reloading support, which should
be accessible under <http://127.0.0.1:8000/>.

    uv run mkdocs serve

Run the following command to perform some basic linting before pushing your
changes to GitHub:

    uv run pre-commit run --all-files

The previous command uses `markdownlint-cli` to lint the docs using a library
of [rules](https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md).
Please follow the link to troubleshoot any linting issues.

## Dependency maintenance

The documentation dependencies are declared in `pyproject.toml` and pinned in
`uv.lock`. Dependabot maintains them automatically. To update them manually,
first preview the changes without modifying the lockfile:

    uv lock --upgrade --dry-run

When the proposed changes are acceptable, update the lockfile and synchronize
the local environment:

    uv lock --upgrade
    uv sync

Before committing the result, verify that the lockfile is current and build the
documentation:

    uv lock --check
    uv run mkdocs build --strict
