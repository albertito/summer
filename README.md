
# summer 🌞 🏖

Utility to detect accidental data corruption (e.g. bitrot, storage media
problems).  Not intended to detect malicious modification.

Checksums are written to/read from each file's extended attributes.


## Status

[![tests](https://github.com/albertito/summer/actions/workflows/tests.yaml/badge.svg)](https://github.com/albertito/summer/actions/workflows/tests.yaml)
[![codecov](https://codecov.io/gh/albertito/summer/graph/badge.svg?token=Nd3STeoyuk)](https://codecov.io/gh/albertito/summer)


summer is still under active development. The user interface and on-disk
format may change in backwards-incompatible ways.


## Install

```
go install blitiri.com.ar/go/summer@latest
```


## Example

The most common use case is to run `summer update`, which writes checksums for
new or modified files, and verifies the checksums of the files which have not
been modified.

The `-x` flag stops summer from crossing filesystem boundaries (for each of
the given paths).

```
sudo summer -x update /home /etc /usr
```

## Contact

If you have any bug reports, questions, comments or patches please send them
to summer@alb.ar.

