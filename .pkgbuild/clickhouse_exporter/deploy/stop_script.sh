#!/bin/bash
# Name    : stop_script.py
# Date    : 2016.03.28
# Func    : 停止脚本
# Note    : 注意：当前路径为应用部署文件夹

#############################################################
# 用户自定义
process_name="clickhouse_exporter"       # 进程名

ps -ef | grep ${process_name} | awk '{print $2}' | xargs kill 2>/dev/null

exit 0
