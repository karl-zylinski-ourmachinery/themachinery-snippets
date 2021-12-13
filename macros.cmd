@echo off

DOSKEY om=cd c:\work\ourmachinery.com $T$T title ourmachinery.com
DOSKEY tm=cd c:\work\themachinery $T$T title The Machinery
DOSKEY debug=c:\work\themachinery\bin\debug\the-machinery.exe $*
DOSKEY ls=dir /w
DOSKEY vscmd="C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvars64.bat"