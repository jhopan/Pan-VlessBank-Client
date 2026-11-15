@echo off
setlocal EnableExtensions EnableDelayedExpansion
set "REPO=%VLESSBANK_CLIENT_REPO%"
if "%REPO%"=="" set "REPO=jhopan/Pan-VlessBank-Client"
set "RELEASE=%VLESSBANK_CLIENT_RELEASE%"
if "%RELEASE%"=="" set "RELEASE=latest"
set "APP_DIR=%VLESSBANK_CLIENT_DIR%"
if "%APP_DIR%"=="" set "APP_DIR=%USERPROFILE%\pan-vlessbank-client"
set "ASSET=pan-vlessbank-client_windows_amd64.exe"
set "BASE_URL=https://github.com/%REPO%/releases/download/%RELEASE%"
set "TMP=%TEMP%\pan-vlessbank-client-install-%RANDOM%%RANDOM%"
mkdir "%TMP%" || (echo ERROR: cannot create temp directory& exit /b 1)
curl.exe -fsSL --retry 3 -o "%TMP%\%ASSET%" "%BASE_URL%/%ASSET%" || (echo ERROR: binary download failed& rmdir /S /Q "%TMP%"& exit /b 1)
curl.exe -fsSL --retry 3 -o "%TMP%\checksums.txt" "%BASE_URL%/checksums.txt" || (echo ERROR: checksum download failed& rmdir /S /Q "%TMP%"& exit /b 1)
for /f "tokens=1" %%A in ('findstr /I /C:"%ASSET%" "%TMP%\checksums.txt"') do set "EXPECTED=%%A"
if "%EXPECTED%"=="" (echo ERROR: checksum not found& rmdir /S /Q "%TMP%"& exit /b 1)
for /f %%A in ('powershell -NoProfile -Command "(Get-FileHash -Algorithm SHA256 -LiteralPath '%TMP%\%ASSET%').Hash"') do set "ACTUAL=%%A"
if /I not "%EXPECTED%"=="%ACTUAL%" (echo ERROR: checksum mismatch& rmdir /S /Q "%TMP%"& exit /b 1)
mkdir "%APP_DIR%" 2>nul
move /Y "%TMP%\%ASSET%" "%APP_DIR%\pan-vlessbank-client.exe" >nul || (echo ERROR: binary replacement failed& rmdir /S /Q "%TMP%"& exit /b 1)
rmdir /S /Q "%TMP%"
echo Installed: %APP_DIR%\pan-vlessbank-client.exe
echo Release: %RELEASE%
echo client.settings.json was preserved. Install cloudflared separately or configure its path in client panel.
echo Run: "%APP_DIR%\pan-vlessbank-client.exe"
