@echo off
setlocal enabledelayedexpansion

if "%~2"=="" (
    echo Usage: %0 ^<provider^> ^<secret_key^>
    echo Supported providers: EOL'
echo -n $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$/\n/') >> gen/store_key.bat
cat >> gen/store_key.bat << 'EOL'
    exit /b 1
)

set "provider=%~1"
set "secret_key=%~2"

:: Convert provider to lowercase
for %%i in (a b c d e f g h i j k l m n o p q r s t u v w x y z) do call set "provider=!provider:%%i=%%i!"

:: Validate provider
set "valid="
for %%p in (EOL'
echo -n $(echo "$PROVIDERS" | tr '\n' ' ') >> gen/store_key.bat
cat >> gen/store_key.bat << 'EOL'
) do (
    if /i "%provider%"=="%%p" set "valid=1"
)

if not defined valid (
    echo Error: Invalid provider. Supported providers are: EOL'
echo -n $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$/\n/') >> gen/store_key.bat
cat >> gen/store_key.bat << 'EOL'
    exit /b 1
)

:: Convert provider to uppercase for environment variable
set "upper_provider=%provider%"
for %%i in (a b c d e f g h i j k l m n o p q r s t u v w x y z) do call set "upper_provider=!upper_provider:%%i=%%i!"
set "env_var=%upper_provider%_API_KEY"

:: Set the environment variable for the current user
setx "%env_var%" "%secret_key%" >nul

:: Also set for current session
set "%env_var%=%secret_key%"

echo Successfully stored API key for %provider%
echo Environment variable %env_var% has been set
echo The environment variable is available system-wide
echo Please restart any applications that need to use this environment variable
