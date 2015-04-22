README
======

You ever wanted to use `uniq` on a line delimited JSON file? You've come to the right place.

Breaking up the problem
-----------------------

When working with large LDJ files, it is inconvenient to store *seen* values
in a *set* because of the linear memory requirements. The traditional `uniq`
is efficient, since it expects sorted input. The first problem therefore would
be to sort a line-delimited JSON file by the values of a certain field.

There is already a fine `sort` on most Unix systems, which is multicore aware since 8.6:

> As of coreutils 8.6 (2010-10-15), GNU sort already sorts in parallel to make use of several processors where available.

From: http://unix.stackexchange.com/a/88704/376

We can bracket the `sort`, so it works with LDJ files, too: First extract the interesting value along with document
boundaries from the LDJ, then sort by the value and then permute the original file, given the sorted boundaries:

Step by step
------------

    $ cat fixtures/test.ldj
    {"name": "Ann", "city": "London"}
    {"name": "Bob", "city": "Paris"}
    {"name": "Claude", "city": "Berlin"}
    {"name": "Diane", "city": "New York"}
    {"name": "Ann", "city": "Moscow"}

    $ lloyd-map -key name fixtures/test.ldj
    Ann 0   34
    Bob 34  33
    Claude  67  37
    Diane   104 38
    Ann 142 34

    $ lloyd-map -key name fixtures/test.ldj | sort
    Ann 0   34
    Ann 142 34
    Bob 34  33
    Claude  67  37
    Diane   104 38

    $ lloyd-map -key name fixtures/test.ldj | sort | cut -f 2-
    0   34
    142 34
    34  33
    67  37
    104 38

This cryptic list contains the document boundaries in order of the sorted values.

    $ lloyd-map -key name fixtures/test.ldj | sort | cut -f 2- | lloyd-permute fixtures/test.ldj
    {"name": "Ann", "city": "London"}
    {"name": "Ann", "city": "Moscow"}
    {"name": "Bob", "city": "Paris"}
    {"name": "Claude", "city": "Berlin"}
    {"name": "Diane", "city": "New York"}

Now it is possible to deduplicate with constant memory:

    $ lloyd-map -key name fixtures/test.ldj | sort | cut -f 2- | lloyd-permute fixtures/test.ldj | lloyd-uniq -key name
    {"name": "Ann", "city": "Moscow"}
    {"name": "Bob", "city": "Paris"}
    {"name": "Claude", "city": "Berlin"}
    {"name": "Diane", "city": "New York"}
