#!/bin/bash
# 停止所有 proxy 进程
#echo "Killing all proxy processes..."
#killall proxy

# 停止所有 nodelet 进程
echo "Killing all nodelet processes..."
killall nodelet

# 停止所有 scheduler 进程
echo "Killing all scheduler processes..."
killall scheduler

# 停止所有 wasm runtime 进程
echo "Killing all wasm runtime processes..."
killall server

echo "All specified processes have been killed."


