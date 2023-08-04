Tests for excluding files.

  $ alias summer="$TESTDIR/../summer"

Simple test data.

  $ touch empty1 empty2 empty3
  $ echo marola > hola

Generate.

  $ summer -n --exclude empty2 --excludere emp..3 generate .
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

Use a bit more complex test data.

  $ mkdir dir1 dir2
  $ touch dir1/file1 dir1/file2 dir1/file3
  $ touch dir2/file1 dir2/file2 dir2/file3

  $ summer -n \
  >   --exclude empty2 \
  >   --excludere emp..3 \
  >   -exclude dir2 \
  >   -excludere d..1/f...2 \
  >   generate .
  0s: 0 matched, 0 modified, 4 new, 0 corrupted
