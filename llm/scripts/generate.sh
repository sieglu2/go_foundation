#!/bin/bash

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed."
    echo "Please install jq first:"
    echo "  For Ubuntu/Debian: sudo apt-get install jq"
    echo "  For MacOS: brew install jq"
    exit 1
fi

# Create gen directory
mkdir -p gen

# Read providers from JSON
PROVIDERS=$(jq -r 'keys[]' ./provider.json)

# Generate delete_all_keys.bat
cat > gen/delete_all_keys.bat << 'EOL'
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
EOL

# Generate delete_all_keys.sh
cat > gen/delete_all_keys.sh << 'EOL'
#!/bin/bash

# Function to delete a credential
delete_credential() {
    local account=$1
    local service=$2
    local provider=$3

    echo "Deleting key for $provider..."
    if security delete-generic-password -a "$account" -s "$service" 2>/dev/null; then
        echo "✓ Successfully deleted API key for $provider"
    else
        echo "✗ Failed to delete API key for $provider (key might not exist)"
    fi
}

EOL

# Add delete commands for each provider
while read -r provider; do
    account_name=$(jq -r ".[\"$provider\"].account_name" ./provider.json)
    service_name=$(jq -r ".[\"$provider\"].service_name" ./provider.json)
    echo "# Delete $provider credentials" >> gen/delete_all_keys.sh
    echo "delete_credential \"$account_name\" \"$service_name\" \"$provider\"" >> gen/delete_all_keys.sh
    echo "" >> gen/delete_all_keys.sh
done <<< "$PROVIDERS"

echo 'echo "Operation completed"' >> gen/delete_all_keys.sh

# Generate store_key.bat
cat > gen/store_key.bat << 'EOL'
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
EOL

# Generate store_key.sh
cat > gen/store_key.sh << EOL
#!/bin/bash

if [ "\$#" -ne 2 ]; then
    echo "Usage: \$0 <provider> <secret_key>"
    echo "Supported providers: $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$//')"
    exit 1
fi

provider=\$1
secret_key=\$2

case "\$provider" in
EOL

# Add case statements for each provider
while read -r provider; do
    account_name=$(jq -r ".[\"$provider\"].account_name" ./provider.json)
    service_name=$(jq -r ".[\"$provider\"].service_name" ./provider.json)
    cat >> gen/store_key.sh << EOL
    "$provider")
        account_name="$account_name"
        service_name="$service_name"
        ;;
EOL
done <<< "$PROVIDERS"

# Add the default case and remaining script
cat >> gen/store_key.sh << EOL
    *)
        echo "Error: Invalid provider. Supported providers: $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$//')"
        exit 1
        ;;
esac

security add-generic-password -U -a "\$account_name" -s "\$service_name" -w "\$secret_key"

echo "Successfully stored API key for \$provider"
EOL

# Generate test_key.bat
cat > gen/test_key.bat << 'EOL'
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
EOL

# Generate test_key.sh
cat > gen/test_key.sh << EOL
#!/bin/bash

if [ "\$#" -ne 1 ]; then
    echo "Usage: \$0 <provider>"
    echo "Supported providers: $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$//')"
    exit 1
fi

provider=\$1

case "\$provider" in
EOL

# Add case statements for each provider
while read -r provider; do
    account_name=$(jq -r ".[\"$provider\"].account_name" ./provider.json)
    service_name=$(jq -r ".[\"$provider\"].service_name" ./provider.json)
    cat >> gen/test_key.sh << EOL
    "$provider")
        account_name="$account_name"
        service_name="$service_name"
        ;;
EOL
done <<< "$PROVIDERS"

# Add the default case and remaining script
cat >> gen/test_key.sh << EOL
    *)
        echo "Error: Invalid provider. Supported providers: $(echo "$PROVIDERS" | tr '\n' ', ' | sed 's/,$//')"
        exit 1
        ;;
esac

if security find-generic-password -a "\$account_name" -s "\$service_name" >/dev/null 2>&1; then
    echo "API key exists for \$provider"
    exit 0
else
    echo "No API key found for \$provider"
    exit 1
fi
EOL

# Make shell scripts executable
chmod +x gen/delete_all_keys.sh gen/store_key.sh gen/test_key.sh

echo "Generated files in 'gen' directory:"
echo "- delete_all_keys.bat"
echo "- delete_all_keys.sh"
echo "- store_key.bat"
echo "- store_key.sh"
echo "- test_key.bat"
echo "- test_key.sh"
