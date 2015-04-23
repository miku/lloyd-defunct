README
======

Did you ever wanted to use `uniq` on a line delimited JSON file? You've come to the right place.

Installation
------------

Go get all utils:

    $ go get github.com/miku/lloyd/cmd/...

Breaking up the problem
-----------------------

When working with large[1] LDJ files, it is inconvenient to store *seen*
values in a *set* because of the linear memory requirements. [Bloom
filters](http://en.wikipedia.org/wiki/Bloom_filter) are more space efficent,
but they allow false positives.

The traditional `uniq` is efficient, since it works on sorted input. The first
problem therefore would be to sort a line-delimited JSON file by a key or keys.

There is already `sort` on most Unix systems, which is multicore aware since 8.6:

> As of coreutils 8.6 (2010-10-15), GNU sort already sorts in parallel to make use of several processors where available.

See also: http://unix.stackexchange.com/a/88704/376 and [9face836f3](http://git.savannah.gnu.org/cgit/coreutils.git/commit/?id=9face836f36c507f01a7d7a33138c5a303e3b1df).

We can bracket the `sort`, so it works with LDJ files, too: First *extract* the interesting value along with document
boundaries from the LDJ, then sort by the value and then *permute* the original file, given the sorted boundaries:

[1] large: does not fit in memory

Step by step
------------

    $ cat fixtures/test.ldj
    {"name": "Ann", "more": {"city": "London", "syno": 4}}
    {"name": "涛", "more": {"city": "香港", "syno": 1}}
    {"name": "Bob", "more": {"city": "Paris", "syno": 3}}
    {"name": "Claude", "more": {"city": "Berlin", "syno": 5}}
    {"name": "Diane", "more": {"city": "New York", "syno": 6}}
    {"name": "Ann", "more": {"city": "Moscow", "syno": 2}}


    $ lloyd-map -keys 'name, more.syno' fixtures/test.ldj
    Ann 4   0   55
    涛   1   55  55
    Bob 3   110 54
    Claude  5   164 58
    Diane   6   222 59
    Ann 2   281 55

    $ lloyd-map -keys 'name, more.syno' fixtures/test.ldj | sort
    Ann 2   281 55
    Ann 4   0   55
    Bob 3   110 54
    Claude  5   164 58
    Diane   6   222 59
    涛   1   55  55

Now a `sort -u` will do the job, if restricted to the first column:

    $ lloyd-map -keys 'name, more.syno' fixtures/test.ldj | sort -uk1,1
    Ann 4   0   55
    Bob 3   110 54
    Claude  5   164 58
    Diane   6   222 59
    涛   1   55  55

Now we only need to *seek and read* to the locations given as offset and
length in the *last two columns* and slice out the corresponding records from
the original file:

    $ lloyd-map -keys 'name, more.syno' fixtures/test.ldj | sort -uk1,1 | cut -f3-
    0   55
    110 54
    164 58
    222 59
    55  55

    $ lloyd-map -keys 'name, more.syno' fixtures/test.ldj | sort -uk1,1 | cut -f3- | lloyd-permute fixtures/test.ldj

    {"name": "Ann", "more": {"city": "London", "syno": 4}}
    {"name": "Bob", "more": {"city": "Paris", "syno": 3}}
    {"name": "Claude", "more": {"city": "Berlin", "syno": 5}}
    {"name": "Diane", "more": {"city": "New York", "syno": 6}}
    {"name": "涛", "more": {"city": "香港", "syno": 1}}

Caveats
-------

Current limitations:

* Only top-level keys are supported yet.
* Only a single key can be specified.
* The values should not contain tabs, since `lloyd-map` currently outputs tab delimited lists.

