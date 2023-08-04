
  $ alias summer="$TESTDIR/../summer"


No arguments.

  $ summer
  # summer 🌞 🏖
  
  Utility to detect accidental data corruption (e.g. bitrot, storage media
  problems).  Not intended to detect malicious modification.
  
  Checksums are written to/read from each files' extended attributes by default,
  or to a separate database file (with the -db flag).
  
  Usage:
  
    summer update <dir>
        Verify checksums in the given directory, and update them for new or
        changed files.
    summer verify <dir>
        Verify checksums in the given directory.
    summer generate <dir>
        Write checksums for the given directory. Files with pre-existing
        checksums are left untouched, and checksums are not verified.
        Useful when generating checksums for a lot of files for the first time,
        as is faster to resume work if interrupted.
    summer version
        Print software version information.
  
  Flags:
    -db string
      \tdatabase to read from/write to (esc)
    -exclude value
      \texclude these paths (can be repeated) (esc)
    -excludere value
      \texclude paths matching this regexp (can be repeated) (esc)
    -forcetty
      \tforce TTY output (esc)
    -n\tdry-run mode (do not write anything) (esc)
    -q\tquiet mode (esc)
    -v\tverbose mode (list each file) (esc)
    -x\tdon't cross filesystem boundaries (esc)
  [1]


Too few arguments.

  $ summer lskfmsl
  # summer 🌞 🏖
  
  Utility to detect accidental data corruption (e.g. bitrot, storage media
  problems).  Not intended to detect malicious modification.
  
  Checksums are written to/read from each files' extended attributes by default,
  or to a separate database file (with the -db flag).
  
  Usage:
  
    summer update <dir>
        Verify checksums in the given directory, and update them for new or
        changed files.
    summer verify <dir>
        Verify checksums in the given directory.
    summer generate <dir>
        Write checksums for the given directory. Files with pre-existing
        checksums are left untouched, and checksums are not verified.
        Useful when generating checksums for a lot of files for the first time,
        as is faster to resume work if interrupted.
    summer version
        Print software version information.
  
  Flags:
    -db string
      \tdatabase to read from/write to (esc)
    -exclude value
      \texclude these paths (can be repeated) (esc)
    -excludere value
      \texclude paths matching this regexp (can be repeated) (esc)
    -forcetty
      \tforce TTY output (esc)
    -n\tdry-run mode (do not write anything) (esc)
    -q\tquiet mode (esc)
    -v\tverbose mode (list each file) (esc)
    -x\tdon't cross filesystem boundaries (esc)
  [1]


No valid path (the argument is given, but it is empty).

  $ summer weifmws ""
  # summer 🌞 🏖
  
  Utility to detect accidental data corruption (e.g. bitrot, storage media
  problems).  Not intended to detect malicious modification.
  
  Checksums are written to/read from each files' extended attributes by default,
  or to a separate database file (with the -db flag).
  
  Usage:
  
    summer update <dir>
        Verify checksums in the given directory, and update them for new or
        changed files.
    summer verify <dir>
        Verify checksums in the given directory.
    summer generate <dir>
        Write checksums for the given directory. Files with pre-existing
        checksums are left untouched, and checksums are not verified.
        Useful when generating checksums for a lot of files for the first time,
        as is faster to resume work if interrupted.
    summer version
        Print software version information.
  
  Flags:
    -db string
      \tdatabase to read from/write to (esc)
    -exclude value
      \texclude these paths (can be repeated) (esc)
    -excludere value
      \texclude paths matching this regexp (can be repeated) (esc)
    -forcetty
      \tforce TTY output (esc)
    -n\tdry-run mode (do not write anything) (esc)
    -q\tquiet mode (esc)
    -v\tverbose mode (list each file) (esc)
    -x\tdon't cross filesystem boundaries (esc)
  [1]


Unknown command.

  $ summer badcommand .
  unknown command "badcommand"
  [1]

Version information.

  $ summer version
  summer version \w+ \(....-..-.. ..:..:.. .*\) (re)
