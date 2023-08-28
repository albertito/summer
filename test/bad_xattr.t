Test how a bad/invalid xattr value is handled. This shouldn't happen normally,
but it might if the xattr itself gets corrupted.

  $ alias summer="$TESTDIR/../summer"
  $ echo marola > hola
  $ summer generate .
  0s: 0 matched, 0 modified, 1 new, 0 corrupted

Corrupt the xattr by writing data that does not serialize into ChecksumV1. We
achieve that by having less data than it expects.

  $ xattr -w user.summer-v1 "xxxx" hola

Verify and check the error.

  $ summer verify .
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
  error in "hola": unexpected EOF
  [1]
