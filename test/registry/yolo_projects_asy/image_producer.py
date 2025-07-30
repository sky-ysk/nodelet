import os
import time
import cv2
import numpy as np
import redis
import base64

def upload_local_images(image_dir, redis_host='localhost', redis_port=6379):
    """上传本地图片到Redis Stream"""
    redis_manager = RedisStreamManager(host=redis_host, port=redis_port)

    # 获取所有图片文件
    image_files = [f for f in os.listdir(image_dir)
                   if f.lower().endswith(('.png', '.jpg', '.jpeg', '.bmp'))]

    print(f"📁 发现 {len(image_files)} 张图片")

    for i, filename in enumerate(image_files):
        image_path = os.path.join(image_dir, filename)

        try:
            # 读取图片
            img = cv2.imread(image_path)
            if img is None:
                print(f"⚠️ 无法读取图片: {filename}")
                continue

            # 添加到Stream
            message_id = redis_manager.add_image(f"img-{i + 1}", img)
            print(f"✅ 上传图片: {filename} -> {message_id}")

            time.sleep(0.1)  # 避免发送过快

        except Exception as e:
            print(f"❌ 上传失败: {filename}, {str(e)}")

    print("🏁 所有图片上传完成")


class RedisStreamManager:
    """Redis Streams管理器"""

    def __init__(self, host='localhost', port=6379, stream_name='image_stream'):
        self.redis = redis.Redis(host=host, port=port)
        self.stream_name = stream_name
        self.group_name = 'inference_group'
        self.consumer_prefix = 'consumer'

        # 创建消费者组 - 修复爆红问题
        try:
            # 使用正确的参数顺序和命名
            self.redis.xgroup_create(
                name=self.stream_name,
                groupname=self.group_name,
                id='0',
                mkstream=True
            )
        except redis.exceptions.ResponseError as e:
            if "BUSYGROUP" in str(e):
                print(f"消费者组 '{self.group_name}' 已存在")
            else:
                raise

    def add_image(self, image_id, image_data):
        """添加图片到Stream"""
        # 编码图片数据
        if isinstance(image_data, np.ndarray):
            _, img_encoded = cv2.imencode('.jpg', image_data)
            image_bytes = img_encoded.tobytes()
        else:
            image_bytes = image_data

        image_base64 = base64.b64encode(image_bytes).decode('utf-8')

        # 创建消息
        message = {
            'image_id': image_id,
            'image_data': image_base64
        }

        # 添加到Stream
        return self.redis.xadd(self.stream_name, message)

    def get_next_image(self, consumer_id):
        """获取下一张图片"""
        # 读取消息
        messages = self.redis.xreadgroup(
            self.group_name, consumer_id,
            {self.stream_name: '>'},
            count=1, block=1000
        )

        if not messages:
            return None, None, None

        # 解析消息
        stream, message_list = messages[0]
        message_id = message_list[0][0].decode('utf-8')
        message_data = message_list[0][1]

        # 提取数据
        image_id = message_data[b'image_id'].decode('utf-8')
        image_base64 = message_data[b'image_data'].decode('utf-8')

        # 解码图片
        image_bytes = base64.b64decode(image_base64)
        nparr = np.frombuffer(image_bytes, np.uint8)
        image = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

        return message_id, image_id, image

    def ack_message(self, message_id):
        """确认消息处理完成"""
        self.redis.xack(self.stream_name, self.group_name, message_id)

    def get_last_message_id(self):
        """获取最后一条消息ID"""
        last_id = self.redis.xrevrange(self.stream_name, count=1)
        if last_id:
            return last_id[0][0].decode('utf-8')
        return None

    def get_pending_messages(self):
        """获取待处理消息"""
        return self.redis.xpending(self.stream_name, self.group_name)

if __name__ == "__main__":
    # 配置参数
    IMAGE_DIR = "./test_input"  # 图片目录
    REDIS_HOST = "localhost"
    REDIS_PORT = 6379

    upload_local_images(IMAGE_DIR, REDIS_HOST, REDIS_PORT)
