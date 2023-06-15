# Release

## Start a new development cycle

Edit [VERSION.txt](/VERSION.txt) (e.g. change "0.1.0" to "0.2.0"), commit and
push the change, e.g.:

    $ echo "0.2.0" > VERSION.txt
    $ git commit -m "Bump version to 0.2.0" VERSION.txt
    $ git push

## Publish git tag

In release day, publish a new git tag matching the version number in
`VERSION.txt`, including the "v" prefix, e.g.:

    $ git tag v0.2.0
    $ git push origin refs/tags/v0.2.0
