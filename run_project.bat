@echo off
REM run_project.bat — quick demo (Windows)
REM - automatically requests admin privileges to kill processes
REM - compiles Go backend (uses compile_golang.bat)
REM - installs frontend deps if missing
REM - kills existing processes on ports 8080 and 3000
REM - launches backend exe and Next.js dev server in separate terminals

REM Check for admin privileges and re-launch if needed
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [sotekre] requesting administrator privileges...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

setlocal
echo [sotekre] preparing demo environment...

REM Kill all existing sotekre.exe processes (backend)
echo [sotekre] killing existing backend processes...
taskkill /F /IM sotekre.exe >nul 2>&1

REM Kill existing processes on port 8080 (backend)
echo [sotekre] checking port 8080...
for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":8080"') do (
    taskkill /F /PID %%a >nul 2>&1
)

REM Kill existing processes on port 3000 (frontend - more aggressive)
echo [sotekre] checking port 3000...
for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":3000"') do (
    echo [sotekre] killing PID %%a on port 3000
    taskkill /F /PID %%a >nul 2>&1
)

REM Also kill any node processes that might be lingering
echo [sotekre] cleaning up node processes...
for /f "tokens=2" %%a in ('tasklist ^| findstr "node.exe"') do (
    taskkill /F /PID %%a >nul 2>&1
)

REM Wait longer for ports to be fully released
echo [sotekre] waiting for ports to be released...
timeout /t 3 /nobreak >nul

REM Verify ports are free
:check_ports
netstat -aon | findstr ":3000" | findstr "LISTENING" >nul 2>&1
if %errorLevel% equ 0 (
    echo [sotekre] port 3000 still in use, waiting...
    timeout /t 2 /nobreak >nul
    goto check_ports
)

REM ensure backend env exists
if not exist "%~dp0backend\.env" (
  copy "%~dp0backend\.env.example" "%~dp0backend\.env" >nul
  echo [sotekre] created backend\.env from example — edit credentials if needed
)

REM compile backend
call "%~dp0compile_golang.bat"
if errorlevel 1 (
  echo [sotekre] backend build failed — aborting
  exit /b 1
)

REM check Node
where node >nul 2>&1
if errorlevel 1 (
  echo ERROR: Node.js not found on PATH. Install Node 18+ from https://nodejs.org/
  exit /b 1
)

REM install frontend deps if needed
if not exist "%~dp0frontend\node_modules" (
  echo [sotekre] installing frontend dependencies - frontend/ ...
  pushd "%~dp0frontend"
  npm install
  if errorlevel 1 (
    echo [sotekre] npm install failed
    popd
    exit /b 1
  )
  popd
)

REM start backend and frontend in separate windows
start "Sotekre - Backend" /D "%~dp0backend" cmd /k .\sotekre.exe
start "Sotekre - Frontend" /D "%~dp0frontend" cmd /k npm run dev

echo
echo [sotekre] launched backend and frontend.
echo Open http://localhost:3000 for the UI and http://localhost:8080/api/menus for the API.
endlocal
pause
