@echo off
echo ========================================================
echo Running Orynt Golang Backend Tests...
echo ========================================================
cd backend
go test -v ./...
if %errorlevel% neq 0 (
    echo [ERROR] Backend tests failed!
    exit /b %errorlevel%
)

echo.
echo ========================================================
echo Running Orynt React/TypeScript Frontend Tests...
echo ========================================================
cd ../frontend
call npm.cmd run test
if %errorlevel% neq 0 (
    echo [ERROR] Frontend tests failed!
    exit /b %errorlevel%
)

echo.
echo ========================================================
echo [SUCCESS] All Orynt backend and frontend tests passed successfully!
echo ========================================================
cd ..
