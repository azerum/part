# Algorithm for crash-safe update of manifest

Goal:

Update file F, such that, if anything crashes mid-update:

- File F is said to be "correct"
- Half-written file (not identical to F) is "incorrect"

- Directory structure after crash clearly defines which files are
correct

- Users can remove incorrect files and re-run
