@echo off
chcp 65001 >nul
echo 正在启动股票数据分析系统...
echo.

echo 启动数据获取服务...
start "数据获取服务" cmd /k "python cron.py"

echo 等待2秒让数据服务先启动...
timeout /t 2 /nobreak >nul

echo 启动Web界面...
start "Web界面" cmd /k "streamlit run Stock.py"

echo.
echo 系统启动完成！
echo 数据获取服务和Web界面已在后台运行
echo 浏览器将自动打开 http://localhost:8501
echo.
echo 按任意键关闭此窗口...
pause >nul