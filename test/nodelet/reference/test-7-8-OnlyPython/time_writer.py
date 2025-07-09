import time

def main():
    # 模拟写入时间戳到文件
    with open("time.txt", "w") as f:
        for i in range(10):
            f.write(f"Time: {time.time()}\n")
            time.sleep(1)  # 模拟每秒写入一次

if __name__ == "__main__":
    main()