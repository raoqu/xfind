ifeq ($(OS),Windows_NT)
	TARGET = xfind.exe
	ARCHIVE = mv $(TARGET) d:\apps\_tools
else
	TARGET = xfind
	ARCHIVE = mv $(TARGET) ~/mylab/_tools/ 
endif

all: build archive
build:
	@echo $(OS)
	go build
archive:
	$(ARCHIVE)
run:
# 	@xfind ~/think -name "*.go" -exclude .git -exec print {} -debug 
#	@xfind . -name "util*.go" -exclude .git -exec print {}
#	@xfind . -name "*.java" -exclude .git -exec count -exec countlines
	@$(TARGET) . -name "*.go" -match "SysProcAttr" -exec print {} -exec printmatch
