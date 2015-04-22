README
======

You ever wanted to use `uniq` on a line delimited JSON file? You've come to the right place.

Breaking up the problem
-----------------------

When working with large (RAM+) LDJ files, it is inconvenient to store *seen* values
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
    {"name": "涛", "city": "香港"}
    {"name": "Bob", "city": "Paris"}
    {"name": "Claude", "city": "Berlin"}
    {"name": "Diane", "city": "New York"}
    {"name": "Ann", "city": "Moscow"}

    $ lloyd-map -key name fixtures/test.ldj
    Ann 0   34
    涛   34  34
    Bob 68  33
    Claude  101 37
    Diane   138 38
    Ann 176 34

    $ lloyd-map -key name fixtures/test.ldj | sort
    Ann 0   34
    Ann 176 34
    Bob 68  33
    Claude  101 37
    Diane   138 38
    涛   34  34

    $ lloyd-map -key name fixtures/test.ldj | sort | cut -f 2-
    0   34
    176 34
    68  33
    101 37
    138 38
    34  34

    $ lloyd-map -key name fixtures/test.ldj | sort | cut -f 2- > permutation.txt

This cryptic list contains the document boundaries in order of the sorted values.

    $ cat permutation.txt | lloyd-permute fixtures/test.ldj

    {"name": "Ann", "city": "London"}
    {"name": "Ann", "city": "Moscow"}
    {"name": "Bob", "city": "Paris"}
    {"name": "Claude", "city": "Berlin"}
    {"name": "Diane", "city": "New York"}
    {"name": "涛", "city": "香港"}

Now it is possible to deduplicate with constant memory:

    $ cat permutation.txt | lloyd-permute fixtures/test.ldj | lloyd-uniq -key name

    {"name": "Ann", "city": "Moscow"}
    {"name": "Bob", "city": "Paris"}
    {"name": "Claude", "city": "Berlin"}
    {"name": "Diane", "city": "New York"}
    {"name": "涛", "city": "香港"}
