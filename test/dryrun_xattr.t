Tests for dry-run mode. Similar to the basic ones.

  $ alias summer="$TESTDIR/../summer"

Generate test data.

  $ touch empty
  $ echo marola > hola

Generate and verify.

  $ summer -n generate .
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

  $ summer -n verify .
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

  $ summer -n verify .
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

Now write data for real, so we can test modification.

  $ summer generate .
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

Check handling of new and updated files.

  $ echo trova > nueva
  $ touch empty
  $ summer -n verify .
  \r (no-eol) (esc)
  0s: 1 matched, 1 modified, 1 new, 0 corrupted
  $ summer -n update .
  \r (no-eol) (esc)
  0s: 1 matched, 1 modified, 1 new, 0 corrupted
  $ summer -n verify .
  \r (no-eol) (esc)
  0s: 1 matched, 1 modified, 1 new, 0 corrupted
