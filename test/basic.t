Use the prebuilt summer binary.

  $ alias summer="$TESTDIR/../summer"

Generate test data.

  $ touch empty
  $ echo marola > hola

Generate and verify.

  $ summer generate .
  2 checksums written
  $ summer verify .
  2 matched, 0 modified, 0 new, 0 corrupted

Check handling of new and updated files.

  $ echo trova > nueva
  $ touch empty
  $ summer verify .
  1 matched, 1 modified, 1 new, 0 corrupted
  $ summer update .
  1 matched, 1 modified, 1 new, 0 corrupted
  $ summer verify .
  3 matched, 0 modified, 0 new, 0 corrupted

Corrupt a file by changing its contents without changing the mtime.

  $ OLD_MTIME=`stat -c "%y" hola`
  $ echo sospechoso >> hola
  $ summer verify .
  2 matched, 1 modified, 0 new, 0 corrupted
  $ touch --date="$OLD_MTIME" hola

  $ summer verify .
  "hola": FILE CORRUPTED - expected:239059f6, got:916db13f
  2 matched, 0 modified, 0 new, 1 corrupted
  detected 1 corrupted files
  [1]

Check that "update" also detects the corruption, and doesn't just step over
it.

  $ summer update .
  "hola": FILE CORRUPTED - expected:239059f6, got:916db13f
  2 matched, 0 modified, 0 new, 1 corrupted
  detected 1 corrupted files
  [1]

But "generate" does override it.

  $ summer generate .
  3 checksums written
  $ summer verify .
  3 matched, 0 modified, 0 new, 0 corrupted

Check verbose and quiet.

  $ summer -v verify .
  "empty": match
  "hola": match
  "nueva": match
  3 matched, 0 modified, 0 new, 0 corrupted
  $ summer -q verify .
  $ summer -q generate .
  $ summer -q update .
  $ summer -q verify .

Check that symlinks are ignored.

  $ ln -s hola thisisasymlink
  $ summer -v verify .
  "empty": match
  "hola": match
  "nueva": match
  3 matched, 0 modified, 0 new, 0 corrupted

Check that the root path doesn't confuse us.

  $ summer -v verify $PWD
  "/.*/empty": match (re)
  "/.*/hola": match (re)
  "/.*/nueva": match (re)
  3 matched, 0 modified, 0 new, 0 corrupted
