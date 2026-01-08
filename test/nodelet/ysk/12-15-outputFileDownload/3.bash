#!/bin/bash
# 从 1.txt 生成 2.txt（每行加 +2）
sed 's/$/+3/' 2.txt > 3.txt
echo "处理完成：2.txt → 3.txt"