$ lloyd-uniq -v
0.2.2

$ time lloyd-uniq -key DOI fixtures/crossref.100k.ldj > /dev/null

real    0m16.168s
user    0m14.736s
sys     0m2.238s

----

$ lloyd-uniq -v
0.2.3

$ time lloyd-uniq -key DOI fixtures/crossref.100k.ldj > /dev/null

real    0m11.158s
user    0m9.985s
sys     0m1.097s
