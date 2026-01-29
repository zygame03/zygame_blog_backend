@echo off
echo Compiling for Linux...
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o zBlog cmd/main.go
echo Done. 