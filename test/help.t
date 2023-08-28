
  $ alias summer="$TESTDIR/../summer"


No arguments.

  $ summer
  # summer 🌞 🏖
  
  Utility to detect accidental data corruption (e.g. bitrot, storage media
  problems).  Not intended to detect malicious modification.
  
  Checksums are written to/read from each file's extended attributes.
  
  Paths given can be files or directories. If a directory is given, it is
  processed recursively.
  
  Usage:
  
    summer [flags] update <paths>
        Verify checksums in the given paths, and update them for new or changed
        files.
    summer [flags] verify <paths>
        Verify checksums in the given paths.
    summer [flags] generate <paths>
        Write checksums for the given paths. Files with pre-existing checksums
        are left untouched, and checksums are not verified.
        Useful when generating checksums for a lot of files for the first time,
        as is faster to resume work if interrupted.
    summer [flags] version
        Print software version information.
  
  Flags:
    -exclude value
      \texclude these paths (can be repeated) (esc)
    -excludere value
      \texclude paths matching this regexp (can be repeated) (esc)
    -forcetty
      \tforce TTY output (esc)
    -n\tdry-run mode (do not write anything) (esc)
    -parallel int
      \tnumber of files to process in parallel (0 = number of CPUs) (esc)
    -q\tquiet mode (esc)
    -v\tverbose mode (list each file) (esc)
    -x\tdon't cross filesystem boundaries (esc)
  [1]


Too few arguments.

  $ summer lskfmsl
  # summer 🌞 🏖
  
  Utility to detect accidental data corruption (e.g. bitrot, storage media
  problems).  Not intended to detect malicious modification.
  
  Checksums are written to/read from each file's extended attributes.
  
  Paths given can be files or directories. If a directory is given, it is
  processed recursively.
  
  Usage:
  
    summer [flags] update <paths>
        Verify checksums in the given paths, and update them for new or changed
        files.
    summer [flags] verify <paths>
        Verify checksums in the given paths.
    summer [flags] generate <paths>
        Write checksums for the given paths. Files with pre-existing checksums
        are left untouched, and checksums are not verified.
        Useful when generating checksums for a lot of files for the first time,
        as is faster to resume work if interrupted.
    summer [flags] version
        Print software version information.
  
  Flags:
    -exclude value
      \texclude these paths (can be repeated) (esc)
    -excludere value
      \texclude paths matching this regexp (can be repeated) (esc)
    -forcetty
      \tforce TTY output (esc)
    -n\tdry-run mode (do not write anything) (esc)
    -parallel int
      \tnumber of files to process in parallel (0 = number of CPUs) (esc)
    -q\tquiet mode (esc)
    -v\tverbose mode (list each file) (esc)
    -x\tdon't cross filesystem boundaries (esc)
  [1]


No valid path (the argument is given, but it is empty).

  $ summer verify ""
  0s: 0 matched, 0 modified, 0 new, 0 corrupted
  lstat : no such file or directory
  [1]


Unknown command.

  $ summer badcommand .
  unknown command "badcommand"
  [1]

Version information.

  $ summer version
  summer version \w+ \(....-..-.. ..:..:.. .*\) (re)
