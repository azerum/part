# part hash

## Goal:

Find what files in the dir have their contents modified since the last hashing

## Constraints:

Reading and hashing each file in the dir is slow
    - Therefore, minimize number of file reads we need to perform

## Solution: 

Remember mtime and hash of each file during hashing. If mtime has not changed
since, assume the file contents has not changed either, skip hashing

Reading mtime is faster than reading file contents

If mtime did change, this does not necessarily mean contents did: read file,
hash, compare with previous hash

## Assumptions:

Assumes that, when we check file at times t0 and t1, t1 > t0:
contents at t0 === contents at t1 -> mtime at t0 === mtime at t1 

This is **false** when:

a. A tool is used to modify mtime directly (not via modifying the file). One
can change file, then restore old mtime

b. File is written into without write() syscall: memory-mapped file, block device

c. mtime has limited resolution (e.g. 1s), and file is changes <1s since
the last hashing - mtime will be the same

d. mtime updates asynchronously, not immediately, after writes (see e.g. `lazytime` on
Linux)

e. Due to crash mid-writing, file contents has updated, but `mtime` didn't
(depends on operations order of particular FS)
