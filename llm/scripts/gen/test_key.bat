@echo off
setlocal enabledelayedexpansion

if "%~1"=="" (
    echo Usage: %0 ^<provider^>
    echo Supported providers: EOL'
echo -n $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$/\n/') >> gen/test_key.bat
cat >> gen/test_key.bat << 'EOL'
    exit /b 1
)

set "provider=%~1"

:: Convert provider to lowercase
for %%i in (a b c d e f g h i j k l m n o p q r s t u v w x y z) do call set "provider=!provider:%%i=%%i!"

:: Validate provider
set "valid="
for %%p in (EOL'
echo -n $(echo "$PROVIDERS" | tr '\n' ' ') >> gen/test_key.bat
cat >> gen/test_key.bat << 'EOL'
) do (
    if /i "%provider%"=="%%p" set "valid=1"
)

if not defined valid (
    echo Error: Invalid provider. Supported providers are: EOL'
echo -n $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$/\n/') >> gen/test_key.bat
cat >> gen/test_key.bat << 'EOL'
    exit /b 1
)

:: Convert provider to uppercase for environment variable
set "upper_provider=%provider%"
for %%i in (a b c d e f g h i j k l m n o p q r s t u v w x y z) do call set "upper_provider=!upper_provider:%%i=%%i!"
set "env_var=%upper_provider%_API_KEY"

:: Check if the environment variable exists
reg query "HKCU\Environment" /v "%env_var%" >nul 2>&1
if !errorlevel! equ 0 (
    echo API key exists for %provider%
    exit /b 0
) else (
    echo No API key found for %provider%
    exit /b 1
)
