#!/bin/bash

go run sigs.k8s.io/controller-tools/cmd/controller-gen object paths=types.go

file1="zz_generated.deepcopy.go"  # 自动生成的深拷贝方法
file2="add_deepcopy.go"  # 额外补充的函数
output_file="deepcopy.go"  # 输出文件
insert_line=10 
if [ ! -f "$file1" ]; then
  echo "$file1 不存在!"
  exit 1
fi

if [ ! -f "$file2" ]; then
  echo "$file2 不存在!"
  exit 1
fi


# 头文件替换
old_package='runtime "k8s.io/apimachinery/pkg/runtime"'
new_package='"hit.edu/framework/pkg/apimachinery/runtime"'

sed -i "s|$old_package|$new_package|" "$file1"

function_content=$(awk '/func /, /^$/ {print $0}' "$file2")

head -n $(($insert_line - 1)) "$file1" > "$output_file" 
echo "$function_content" >> "$output_file" 
tail -n +$insert_line "$file1" >> "$output_file" 


rm $file1
#rm $file2
echo "深拷贝方法已保存在$output_file"
