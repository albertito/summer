Test that summer works fine when given a file instead of a directory.

  $ alias summer="$TESTDIR/../summer"
  $ touch empty
  $ echo marola > hola

  $ summer -v generate ./empty
  "./empty": writing checksum \(checksum:0, mtime:\d+\) (re)
  0s: 0 matched, 0 modified, 1 new, 0 corrupted

  $ summer -v verify .
  "empty": match \(checksum:0, mtime:\d+\) (re)
  "hola": missing checksum attribute
  0s: 1 matched, 0 modified, 1 new, 0 corrupted

  $ summer update ./hola
  0s: 0 matched, 0 modified, 1 new, 0 corrupted

  $ summer -v verify .
  "empty": match \(checksum:0, mtime:\d+\) (re)
  "hola": match \(checksum:\w+, mtime:\d+\) (re)
  0s: 2 matched, 0 modified, 0 new, 0 corrupted

