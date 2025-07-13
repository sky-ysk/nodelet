# 功能是读取一个参数，把这个参数写到test.txt内部
import os
import sys

def main():
    if len(sys.argv) < 2:
        print("Usage: python test.py <message>")
        sys.exit(1)

    message = sys.argv[1]
    
    # 写入到test.txt文件
    with open("test.txt", "w") as file:
        file.write(message)
    
    print(f"Message '{message}' written to test.txt")

if __name__ == "__main__":
    main()