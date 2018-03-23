# shallowizer

## what

This script attempts to convert all git repos in the `$GOPATH` to shallow clones in order to save space.
It works as follows:

1. Walk through all directories and sub-directories in `$GOPATH/src`.
2. If a directory contains a `.git` sub-folder, mark it as a repo.
3. Run `git clone --depth=1 "url-of-remote" temporary/directory/` in each repo.
4. Replace the repo with the shallow cloned one.

## why

I have a chromebook running debian via crouton with 3gb of free space.
I can't afford saving commits of repos that I don't even know exist on my machine.

## results

```
$ du --max-depth=0 -h $HOME/go/src
238M	/home/t/go/src
$ shallowizer
???
```

## WARNING

This deletes stuff on your machine. **USE WITH CAUTION**.
