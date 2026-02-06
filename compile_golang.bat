@echo off
REM compile_golang.bat — build the Go backend as a Windows .exe
REM Usage: double-click or run from repository root

setlocal
echo [sotekre] checking Go toolchain...
where go >nul 2>&1
if errorlevel 1 (
  echo ERROR: Go not found on PATH. Install Go (https://go.dev/dl/) and re-run.
  exit /b 1
)

pushd "%~dp0backend" || (echo backend folder not found & exit /b 1)

where swag >nul 2>&1
if errorlevel 1 (
  echo [sotekre] swag CLI not found — docs will not be auto-generated. To generate: go install github.com/swaggo/swag/cmd/swag@latest && go generate ./...
) else (
  echo [sotekre] generating swagger docs...
  go generate ./...
)

echo [sotekre] running: go build -trimpath -ldflags "-s -w" -o sotekre.exe .
set GOFLAGS=
go build -trimpath -ldflags "-s -w" -o sotekre.exe .
if errorlevel 1 (
  echo Build failed.
  popd
  exit /b 1
)

echo [sotekre] built: %CD%\sotekre.exe
popd
endlocal
exit /b 0
