$ErrorActionPreference = 'Stop'
py "$PSScriptRoot/trace.py" @args
exit $LASTEXITCODE
