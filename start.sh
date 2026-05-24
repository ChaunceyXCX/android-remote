#!/bin/bash
cd /home/chauncey/MyProject/android-remote
pkill -9 -f "./android-remote" 2>/dev/null
sleep 1
go build -o android-remote . && nohup ./android-remote > /tmp/android-remote.log 2>&1 &
echo "服务已启动，PID: $!"