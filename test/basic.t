Use the prebuilt summer binary.

  $ alias summer="$TESTDIR/../summer"

Generate test data.

  $ touch empty
  $ echo marola > hola

Generate and verify.

  $ summer generate .
  0s: 0 matched, 0 modified, 2 new, 0 corrupted

  $ summer verify .
  0s: 2 matched, 0 modified, 0 new, 0 corrupted

Check handling of new and updated files.

  $ sleep 0.005
  $ echo trova > nueva
  $ touch empty
  $ summer verify .
  0s: 1 matched, 1 modified, 1 new, 0 corrupted
  $ summer update .
  0s: 1 matched, 1 modified, 1 new, 0 corrupted
  $ summer verify .
  0s: 3 matched, 0 modified, 0 new, 0 corrupted

Corrupt a file by changing its contents without changing the mtime.

  $ OLD_MTIME=`stat -c "%y" hola`
  $ echo sospechoso >> hola
  $ summer verify .
  0s: 2 matched, 1 modified, 0 new, 0 corrupted
  $ touch --date="$OLD_MTIME" hola

  $ summer verify .
  "hola": FILE CORRUPTED - expected:239059f6, got:916db13f
  0s: 2 matched, 0 modified, 0 new, 1 corrupted
  detected 1 corrupted files
  [1]

Check that "update" also detects the corruption, and doesn't just step over
it.

  $ summer update .
  "hola": FILE CORRUPTED - expected:239059f6, got:916db13f
  0s: 2 matched, 0 modified, 0 new, 1 corrupted
  detected 1 corrupted files
  [1]

Editing the file makes us ignore the previous checksum.

  $ touch hola
  $ summer update .
  0s: 2 matched, 1 modified, 0 new, 0 corrupted
  $ summer verify .
  0s: 3 matched, 0 modified, 0 new, 0 corrupted

Check verbose and quiet.

  $ touch denuevo
  $ summer -v verify .
  "denuevo": missing checksum attribute
  "empty": match \(checksum:0, mtime:\d+\) (re)
  "hola": match \(checksum:916db13f, mtime:\d+\) (re)
  "nueva": match \(checksum:91f3a28e, mtime:\d+\) (re)
  0s: 3 matched, 0 modified, 1 new, 0 corrupted
  $ summer -v generate .
  "denuevo": writing checksum \(checksum:0, mtime:\d+\) (re)
  0s: 0 matched, 0 modified, 1 new, 0 corrupted
  $ summer -q verify .
  $ summer -q generate .
  $ summer -q update .
  $ summer -q verify .
  $ rm denuevo

Check that symlinks are ignored.

  $ ln -s hola thisisasymlink
  $ summer -v verify .
  "empty": match \(checksum:0, mtime:\d+\) (re)
  "hola": match \(checksum:916db13f, mtime:\d+\) (re)
  "nueva": match \(checksum:91f3a28e, mtime:\d+\) (re)
  0s: 3 matched, 0 modified, 0 new, 0 corrupted

Check that the root path doesn't confuse us.

  $ summer -v verify $PWD
  "/.*/empty": match \(checksum:0, mtime:\d+\) (re)
  "/.*/hola": match \(checksum:916db13f, mtime:\d+\) (re)
  "/.*/nueva": match \(checksum:91f3a28e, mtime:\d+\) (re)
  0s: 3 matched, 0 modified, 0 new, 0 corrupted
