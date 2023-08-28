
Tests for handling multiple roots on the same command invocation.

  $ alias summer="$TESTDIR/../summer"
  $ mkdir A B C
  $ touch A/a1 A/a2 B/b1 C/c1

  $ summer generate A B C
  0s: 0 matched, 0 modified, 4 new, 0 corrupted
  $ summer update A B C
  0s: 4 matched, 0 modified, 0 new, 0 corrupted
  $ summer verify A B C
  0s: 4 matched, 0 modified, 0 new, 0 corrupted


Test that individual files work well as roots (common use case).

  $ touch B/b2
  $ summer generate A/* B/* C/c1
  0s: 0 matched, 0 modified, 1 new, 0 corrupted
  $ summer update A/* B/* C/c1
  0s: 5 matched, 0 modified, 0 new, 0 corrupted
  $ summer verify A/* B/* C/c1
  0s: 5 matched, 0 modified, 0 new, 0 corrupted


Check the order is as expected (when parallel=1, otherwise the order is not
reproducible).

  $ summer --parallel=1 -v update A B C
  "A/a1": match \(checksum:0, mtime:\d+\) (re)
  "A/a2": match \(checksum:0, mtime:\d+\) (re)
  "B/b1": match \(checksum:0, mtime:\d+\) (re)
  "B/b2": match \(checksum:0, mtime:\d+\) (re)
  "C/c1": match \(checksum:0, mtime:\d+\) (re)
  0s: 5 matched, 0 modified, 0 new, 0 corrupted


Check how we handle getting an error in the middle.

  $ chmod 0000 B/b1

  $ summer --parallel=1 -v verify A B C
  "A/a1": match \(checksum:0, mtime:\d+\) (re)
  "A/a2": match \(checksum:0, mtime:\d+\) (re)
  0s: 2 matched, 0 modified, 0 new, 0 corrupted
  open B/b1: permission denied
  [1]

  $ summer --parallel=1 -v update A B C
  "A/a1": match \(checksum:0, mtime:\d+\) (re)
  "A/a2": match \(checksum:0, mtime:\d+\) (re)
  0s: 2 matched, 0 modified, 0 new, 0 corrupted
  open B/b1: permission denied
  [1]

  $ summer -v generate A B C
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
  open B/b1: permission denied
  [1]
