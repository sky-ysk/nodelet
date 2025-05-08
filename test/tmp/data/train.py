import threading
import time
import os
from ultralytics import YOLO

# 定义一个计时器线程，用于 20 秒后触发退出
def timer_exit(seconds):
    time.sleep(seconds)
    print(f"程序运行时间已达到 {seconds} 秒，自动退出。")
    os._exit(0)  # 强制终止程序

# 训练任务与之前一致
def train_task():
    try:
        model = YOLO("/home/public/workspace/heongtong_yolo_linux/yolo11x-seg.pt")  # 加载预训练模型
        results = model.train(
            data="/home/public/workspace/heongtong_yolo_linux/segment_train.yaml",
            epochs=100,  # 设置较大的 epoch，但程序会在 20 秒后退出
            imgsz=640,
            batch=2,
            device="cpu",
            workers=0
        )
    except Exception as e:
        print(f"训练过程中发生错误: {e}")
    finally:
        print("训练任务已结束。")

if __name__ == '__main__':
    # 启动计时器线程
    timer_thread = threading.Thread(target=timer_exit, args=(20,))
    timer_thread.start()

    # 启动训练任务线程
    train_thread = threading.Thread(target=train_task)
    train_thread.start()

    # 等待两个线程完成
    train_thread.join()
    timer_thread.join()

