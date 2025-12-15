#!/bin/bash
# 从 1.txt 生成 2.txt（每行加 +2）
sed 's/$/+2/' 1.txt > 2.txt
echo "处理完成：1.txt → 2.txt"