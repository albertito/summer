Tests for only working on a subset of the files.

  $ alias summer="$TESTDIR/../summer"

Test data, with some directories (so we can check we go in depth regardless of
the random subset).

  $ mkdir dir1 dir2 dir3
  $ touch dir1/file1 dir1/file2 dir1/file3
  $ touch dir2/file1 dir2/file2 dir2/file3
  $ touch dir3/file1 dir3/file2 dir3/file3

Pick a 50% subset, verify that only the files we expect are considered. Note we
need to use a fixed seed and disable parallelism so we have fully predictable
output to test against.

  $ summer -v \
  >   -subsetseed=69 -subsetpct=50 -parallel=1 \
  >   generate .
  "dir1/file1": writing checksum \(checksum:0, mtime:\d+\) (re)
  "dir1/file2": writing checksum \(checksum:0, mtime:\d+\) (re)
  "dir1/file3": writing checksum \(checksum:0, mtime:\d+\) (re)
  "dir2/file1": writing checksum \(checksum:0, mtime:\d+\) (re)
  "dir3/file2": writing checksum \(checksum:0, mtime:\d+\) (re)
  0s: 0 matched, 0 modified, 5 new, 0 corrupted

Verify using same subset and seed.

  $ summer -v \
  >   -subsetseed=69 -subsetpct=50 -parallel=1 \
  >   verify .
  "dir1/file1": match \(checksum:0, mtime:\d+\) (re)
  "dir1/file2": match \(checksum:0, mtime:\d+\) (re)
  "dir1/file3": match \(checksum:0, mtime:\d+\) (re)
  "dir2/file1": match \(checksum:0, mtime:\d+\) (re)
  "dir3/file2": match \(checksum:0, mtime:\d+\) (re)
  0s: 5 matched, 0 modified, 0 new, 0 corrupted

Check -subset flag validation.

  $ summer -subsetpct=101 verify .
  subset percentage 101 must be in the [0, 100] range
  [1]

Test that with a 100% subset, all files are included.

  $ summer -subsetpct=100 verify .
  0s: 5 matched, 0 modified, 4 new, 0 corrupted

Test that with a 0% subset, no files are included.

  $ summer -subsetpct=0 verify .
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
