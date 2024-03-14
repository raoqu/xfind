# xfind

a cross-platform 'find' alternative.

```shell
xfind

# xfind . -name "util*.go" -exclude .git -exec print {}

# xfind . -name "*.java" -exclude .git -exec count -exec countlines

# xfind . -name "*.go" -match "SysProcAttr" -exec print {} -exec printmatch -exec countmatch

# xfind . -name node_modules -type dir -delete

# xfind ~/think -name "*.go" -exclude .git -exec print {} -debug 
```