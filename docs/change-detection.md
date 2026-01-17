# Algorithm for detecting modified files

Goal:

Find what files in the dir have their contents modified since the last hashing

Constraints:

Reading and hashing each file in the dir is slow
    - Therefore, minimize number of file reads we need to perform

## Algorithm:

Record `manifest: Map<filePath, { mtime, hash }>`

Given existing manifest and a file, tell if it has changed:

1. If file mtime === mtime in manifest: return unchanged
2. Else:
3. Read file contents
4. If contents hash == hash in manifest: return unchanged
5. Else: return changed

## Assumptions 

It makes assumption:
mtime is the same -> (implies) contents are the same

This assumption is *wrong* with local FS when:

a. We change file, than manually restore its mtime

b. We write to file without write() syscall, e.g. memory-mapped file, block device

c. mtime resolution is coarse enough (e.g. 1s), so changes that happen within that
time window are not detected

d. mtime updates are asynchronous after writes, e.g. see `lazytime`

e. Crashes cause contents and metadata to be out of sync

For our usage, this is acceptable:

- We don't run exotic metadata-altering programs and don't access files
via exotic APIs (a, b)

- Files are almost never updated. When they are, updating happens well over 1s
since the last one (c)

- We accept d and e as possible just because of how rare writes are
