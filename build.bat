@echo off
set Dist=dist\

if exist %Dist% (
    rem dist exist
    rmdir /q /s %Dist%
    md %Dist%
) else (
    rem not exist
    md %Dist%
)

SET CGO_ENABLED=0
SET GOARCH=amd64
SET GOOS=linux

echo go build now ...
go build -o server

rem @echo off
echo copy to dist
copy .\server %Dist%
del .\server
rem md %Dist%\Config
rem xcopy .\Config\*.* %Dist%\Config
rem del %Dist%\Config\.gitkeep

cls
echo build on linux ok!
ping -n 6 127.1 >nul