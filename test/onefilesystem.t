
Tests for -x (one filesystem) flag.

We use unshare to create a user+mount namespace, mount a tmpfs inside the test
tree, and verify that -x prevents crossing the boundary.

Skip if we can't create user+mount namespaces (e.g. CI environments).

  $ unshare --user --mount --map-root-user true 2>/dev/null || exit 80

Run the rest of the test inside the namespace, so we can mount a tmpfs.

  $ unshare --user --mount --map-root-user cram3 $TESTDIR/onefilesystem.tt
  .
  # Ran 1 tests, 0 skipped, 0 failed.
