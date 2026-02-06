@echo off
REM run_project.bat — quick demo (Windows)
REM - compiles Go backend (uses compile_golang.bat)
REM - installs frontend deps if missing
REM - launches backend exe and Next.js dev server in separate terminals

setlocal
echo [sotekre] preparing demo environment...

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
