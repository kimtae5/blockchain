@echo on
start "1" cmd /k D:\taeokim\3th\blockchainProject\final\career\Core_pBFT\pBFT\1\consensusPBFT.exe 10000
timeout 3
start "2" cmd /k D:\taeokim\3th\blockchainProject\final\career\Core_pBFT\pBFT\2\consensusPBFT.exe 20000
timeout 3
start "3" cmd /k D:\taeokim\3th\blockchainProject\final\career\Core_pBFT\pBFT\3\consensusPBFT.exe 30000
timeout 3
start "4" cmd /k D:\taeokim\3th\blockchainProject\final\career\Core_pBFT\pBFT\4\consensusPBFT.exe 40000
