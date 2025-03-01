@echo off
setlocal enabledelayedexpansion

:: List of providers
set "providers=EOL'
echo -n $(echo "$PROVIDERS" | tr '\n' ' ') >> gen/delete_all_keys.bat
cat >> gen/delete_all_keys.bat << 'EOL'
"

:: Convert providers to uppercase and clear their environment variables
for %%p in (%providers%) do (
    :: Convert to uppercase for environment variable name
    set "upper_provider=%%p"
    for %%i in (a b c d e f g h i j k l m n o p q r s t u v w x y z) do (
        set "upper_provider=!upper_provider:%%i=%%i!"
    )
    
    set "env_var=!upper_provider!_API_KEY"
    
    :: Check if the environment variable exists
    reg query "HKCU\Environment" /v "!env_var!" >nul 2>&1
    if !errorlevel! equ 0 (
        :: Remove from registry (permanent)
        reg delete "HKCU\Environment" /v "!env_var!" /f >nul 2>&1
        :: Clear from current session
        set "!env_var!="
        echo Cleared API key for %%p
    ) else (
        echo No API key found for %%p
    )
)

echo.
echo All API keys have been cleared
echo Please restart any applications that were using these environment variables
