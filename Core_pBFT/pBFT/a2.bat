@echo on


set /a x=10
:start1
start "%x%000" cmd /k D:\taeokim\3th\blockchainProject\0811jaeho\polygon\1\consensusPBFT.exe %x%000
if %x% == 50 goto exit1
set /a x=x+1
timeout 0.1
goto start1
:exit1

