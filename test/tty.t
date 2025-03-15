Test output to a TTY.

summer will auto-detect if stdout is a tty or not, and change some of the
output accordingly. In this test framework, stdout is not a TTY, so all other
tests use that codepath.

In this test we force tty output, and check the output is as expected.

  $ alias summer="$TESTDIR/../summer"

  $ touch file1 file2 file3

  $ summer -forcetty -n generate .
  \r (no-eol) (esc)
  0s: 0 matched, 0 modified, 3 new, 0 corrupted  
