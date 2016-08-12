sqlitediff
----------
go sqlite databases diff tool

Installation
------------

    go get github.com/vazrupe/sqlitediff


Usage
-----

    import (
        ...
        "github.com/vazrupe/sqlitediff"
        ...
    )

    ...
    d, err := sqlitediff.Diff(YOUR_BEFORE_DBNAME, YOUR_AFTER_DBNAME)
    ...

License
-------
MIT Lisense.