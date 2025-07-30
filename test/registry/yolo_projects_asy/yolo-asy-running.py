import json
import os
import threading
import time
import uuid
from concurrent import futures
import grpc
import cv2
import onnxruntime as ort
import numpy as np
import redis
import base64
from runtime_interface import runtimeIntent_pb2, runtimeIntent_pb2_grpc
import logging
from PIL import Image

# 配置参数
confidence_thres = 0.35
iou_thres = 0.5
classes = {0: 'person', 1: 'bicycle', 2: 'car', 3: 'motorcycle', 4: 'airplane', 5: 'bus', 6: 'train', 7: 'truck',
           8: 'boat', 9: 'traffic light', 10: 'fire hydrant', 11: 'stop sign', 12: 'parking meter', 13: 'bench',
           14: 'bird', 15: 'cat', 16: 'dog', 17: 'horse', 18: 'sheep', 19: 'cow', 20: 'elephant', 21: 'bear',
           22: 'zebra', 23: 'giraffe', 24: 'backpack', 25: 'umbrella', 26: 'handbag', 27: 'tie', 28: 'suitcase',
           29: 'frisbee', 30: 'skis', 31: 'snowboard', 32: 'sports ball', 33: 'kite', 34: 'baseball bat',
           35: 'baseball glove', 36: 'skateboard', 37: 'surfboard', 38: 'tennis racket', 39: 'bottle',
           40: 'wine glass', 41: 'cup', 42: 'fork', 43: 'knife', 44: 'spoon', 45: 'bowl', 46: 'banana', 47: 'apple',
           48: 'sandwich', 49: 'orange', 50: 'broccoli', 51: 'carrot', 52: 'hot dog', 53: 'pizza', 54: 'donut',
           55: 'cake', 56: 'chair', 57: 'couch', 58: 'potted plant', 59: 'bed', 60: 'dining table', 61: 'toilet',
           62: 'tv', 63: 'laptop', 64: 'mouse', 65: 'remote', 66: 'keyboard', 67: 'cell phone', 68: 'microwave',
           69: 'oven', 70: 'toaster', 71: 'sink', 72: 'refrigerator', 73: 'book', 74: 'clock', 75: 'vase',
           76: 'scissors', 77: 'teddy bear', 78: 'hair drier', 79: 'toothbrush'}

color_palette = np.random.uniform(100, 255, size=(len(classes), 3))

providers = [
    # ('CUDAExecutionProvider', {
    #     'device_id': 0,
    # }),
    'CPUExecutionProvider',
]

# Redis配置
REDIS_HOST = os.getenv("REDIS_HOST", "172.110.0.103")  #主要修改连接redis的ip地址  localhost
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))
STREAM_NAME = "image_stream"
GROUP_NAME = "inference_group"

# 全局状态
life_state = -1  # -1:未启动, 0:运行中, 1:暂停中
current_message_id = None
last_processed_id = None
processing_thread = None
session = None
model_inputs = None
input_width = None
input_height = None
redis_manager = None
state_lock = threading.Lock()  # 添加状态锁

# 日志配置
LOG_FORMAT = "%(asctime)s - %(levelname)s - %(message)s"
logging.basicConfig(level=logging.DEBUG, format=LOG_FORMAT)
logger = logging.getLogger(__name__)


def calculate_iou(box, other_boxes):
    x1 = np.maximum(box[0], np.array(other_boxes)[:, 0])
    y1 = np.maximum(box[1], np.array(other_boxes)[:, 1])
    x2 = np.minimum(box[0] + box[2], np.array(other_boxes)[:, 0] + np.array(other_boxes)[:, 2])
    y2 = np.minimum(box[1] + box[3], np.array(other_boxes)[:, 1] + np.array(other_boxes)[:, 3])
    intersection_area = np.maximum(0, x2 - x1) * np.maximum(0, y2 - y1)
    box_area = box[2] * box[3]
    other_boxes_area = np.array(other_boxes)[:, 2] * np.array(other_boxes)[:, 3]
    iou = intersection_area / (box_area + other_boxes_area - intersection_area)
    return iou
def custom_NMSBoxes(boxes, scores, confidence_threshold, iou_threshold):
    if len(boxes) == 0:
        return []
    scores = np.array(scores)
    boxes = np.array(boxes)
    mask = scores > confidence_threshold
    filtered_boxes = boxes[mask]
    filtered_scores = scores[mask]
    if len(filtered_boxes) == 0:
        return []
    sorted_indices = np.argsort(filtered_scores)[::-1]
    indices = []
    while len(sorted_indices) > 0:
        current_index = sorted_indices[0]
        indices.append(current_index)
        if len(sorted_indices) == 1:
            break
        current_box = filtered_boxes[current_index]
        other_boxes = filtered_boxes[sorted_indices[1:]]
        iou = calculate_iou(current_box, other_boxes)
        non_overlapping_indices = np.where(iou <= iou_threshold)[0]
        sorted_indices = sorted_indices[non_overlapping_indices + 1]
    return indices



def draw_detections(img, box, score, class_id):
    x1, y1, w, h = box
    color = color_palette[class_id]
    cv2.rectangle(img, (int(x1), int(y1)), (int(x1 + w), int(y1 + h)), color, 2)
    label = f'{classes[class_id]}: {score:.2f}'
    (label_width, label_height), _ = cv2.getTextSize(label, cv2.FONT_HERSHEY_SIMPLEX, 0.5, 1)
    label_x = x1
    label_y = y1 - 10 if y1 - 10 > label_height else y1 + 10
    cv2.rectangle(img, (label_x, label_y - label_height), (label_x + label_width, label_y + label_height), color,
                  cv2.FILLED)
    cv2.putText(img, label, (label_x, label_y), cv2.FONT_HERSHEY_SIMPLEX, 0.5, (0, 0, 0), 1, cv2.LINE_AA)


def preprocess(img, input_width, input_height):
    img_height, img_width = img.shape[:2]
    img = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
    img = cv2.resize(img, (input_width, input_height))
    image_data = np.array(img) / 255.0
    image_data = np.transpose(image_data, (2, 0, 1))
    image_data = np.expand_dims(image_data, axis=0).astype(np.float32)
    return image_data, img_height, img_width


def postprocess(input_image, output, input_width, input_height, img_width, img_height):
    outputs = np.transpose(np.squeeze(output[0]))
    rows = outputs.shape[0]
    boxes = []
    scores = []
    class_ids = []
    x_factor = img_width / input_width
    y_factor = img_height / input_height
    for i in range(rows):
        classes_scores = outputs[i][4:]
        max_score = np.amax(classes_scores)
        if max_score >= confidence_thres:
            class_id = np.argmax(classes_scores)
            x, y, w, h = outputs[i][0], outputs[i][1], outputs[i][2], outputs[i][3]
            left = int((x - w / 2) * x_factor)
            top = int((y - h / 2) * y_factor)
            width = int(w * x_factor)
            height = int(h * y_factor)
            class_ids.append(class_id)
            scores.append(max_score)
            boxes.append([left, top, width, height])
    indices = custom_NMSBoxes(boxes, scores, confidence_thres, iou_thres)
    for i in indices:
        box = boxes[i]
        score = scores[i]
        class_id = class_ids[i]
        draw_detections(input_image, box, score, class_id)
    return input_image


def init_detect_model(model_path):
    session = ort.InferenceSession(model_path, providers=providers)
    model_inputs = session.get_inputs()
    input_shape = model_inputs[0].shape
    input_width = input_shape[2]
    input_height = input_shape[3]
    return session, model_inputs, input_width, input_height


def detect_object(image, session, model_inputs, input_width, input_height):
    if isinstance(image, Image.Image):
        result_image = np.array(image)
    else:
        result_image = image
    img_data, img_height, img_width = preprocess(result_image, input_width, input_height)
    outputs = session.run(None, {model_inputs[0].name: img_data})
    output_image = postprocess(result_image, outputs, input_width, input_height, img_height, img_width)
    return output_image


def ensure_dir(file_path):
    directory = os.path.dirname(file_path)
    if not os.path.exists(directory):
        os.makedirs(directory)


class RedisStreamManager:
    """Redis Streams管理器（每次创建新的消费者组）"""

    def __init__(self, host=REDIS_HOST, port=REDIS_PORT, stream_name=STREAM_NAME, group_name=None):
        self.redis = redis.Redis(host=host, port=port)
        self.stream_name = stream_name

        # 生成随机消费者组名称（每次创建新组）
        self.group_name = group_name or f"inference-group-{uuid.uuid4().hex}"
        logger.info(f"使用消费者组: {self.group_name}")

        # 创建消费者组（如果不存在）
        try:
            self.redis.xgroup_create(
                name=stream_name,
                groupname=self.group_name,
                id='0',
                mkstream=True
            )
            logger.info(f"创建消费者组: {self.group_name}")
        except redis.exceptions.ResponseError as e:
            if "BUSYGROUP" in str(e):
                logger.info(f"消费者组 '{self.group_name}' 已存在")
            else:
                logger.error(f"创建消费者组失败: {str(e)}")
                raise

    def get_next_image(self, consumer_id):
        """获取下一张图片，支持从指定ID开始"""
        try:
            # 默认从最新消息开始
            messages = self.redis.xreadgroup(
                self.group_name, consumer_id,
                {self.stream_name: '>'},  # 只获取新消息
                count=1, block=10000  # 阻塞10秒，减少空轮询
            )

            # 处理消息
            return self._process_messages(messages)

        except Exception as e:
            logger.error(f"读取消息失败: {str(e)}")
            return None, None, None

    def _process_messages(self, messages):
        """处理消息列表"""
        if not messages:
            # logger.info("没有从Redis获取到消息")
            return None, None, None

        # 添加多层安全检查
        if not messages or not messages[0] or len(messages[0]) < 2:
            logger.info("消息格式无效")
            return None, None, None

        stream, message_list = messages[0]

        # 处理空消息列表的情况
        if not message_list:
            logger.info("消息列表为空")
            return None, None, None

        if not message_list[0] or len(message_list[0]) < 2:
            logger.info("消息列表格式无效")
            return None, None, None

        message_id = message_list[0][0].decode('utf-8')
        message_data = message_list[0][1]

        return self._process_message(message_id, message_data)

    def _process_message(self, message_id, message_data):
        """处理单个消息"""
        # logger.info(f"消息ID: {message_id}")
        # logger.info(f"消息内容: {message_data}")

        # 提取数据
        image_id = message_data.get(b'image_id', b'').decode('utf-8')
        image_base64 = message_data.get(b'image_data', b'').decode('utf-8')

        if not image_id or not image_base64:
            logger.warning(f"消息格式错误: 缺少必要字段")
            return None, None, None

        # logger.info(f"成功获取消息: {image_id} (消息ID: {message_id})")

        # 解码图片
        try:
            image_bytes = base64.b64decode(image_base64)
            nparr = np.frombuffer(image_bytes, np.uint8)
            image = cv2.imdecode(nparr, cv2.IMREAD_COLOR)

            if image is None:
                logger.error("图片解码失败")
                return None, None, None

            return message_id, image_id, image
        except Exception as e:
            logger.error(f"图片解码失败: {str(e)}")
            return None, None, None

    def ack_message(self, message_id):
        """确认消息处理完成"""
        self.redis.xack(self.stream_name, self.group_name, message_id)

    def get_last_message_id(self):
        """获取最后一条消息ID"""
        last_id = self.redis.xrevrange(self.stream_name, count=1)
        if last_id:
            return last_id[0][0].decode('utf-8')
        return None

    def reset_consumer_group(self, id='0-0'):
        """重置消费者组读取位置"""
        try:
            # 检查ID格式是否有效
            if '-' not in id:
                logger.warning(f"无效的消息ID格式: {id}, 使用默认值 '0-0'")
                id = '0-0'

            self.redis.xgroup_setid(self.stream_name, self.group_name, id)
            logger.info(f"重置消费者组 '{self.group_name}' 位置到 {id}")
        except redis.exceptions.ResponseError as e:
            if "NOGROUP" in str(e):
                logger.error(f"消费者组 '{self.group_name}' 不存在")
            elif "invalid ID" in str(e):
                logger.error(f"无效的消息ID: {id}")
            else:
                logger.error(f"重置消费者组失败: {str(e)}")
            raise


class RuntimeService(runtimeIntent_pb2_grpc.RuntimeIntentServicer):
    def __init__(self, task_id):
        self.task_id = task_id
        # 不再需要实例变量，所有状态使用全局变量
        self.state_lock = threading.Lock()

    def init(self, request, context):
        logger.info('-' * 10 + self.__class__.__name__ + '--init()--start' + '-' * 10)
        # 创建新的Redis管理器（新的随机消费者组）
        redis_manager = RedisStreamManager()
        logger.info(f"创建新的Redis管理器，消费者组: {redis_manager.group_name}")
        logger.info('-' * 10 + self.__class__.__name__ + '--init()--end' + '-' * 10)
        return runtimeIntent_pb2.Result(msg='init finished', stateCode=1)

    def start(self, request, context):
        logger.info('-' * 10 + self.__class__.__name__ + '--start()--start' + '-' * 10)
        global life_state, processing_thread, current_message_id, last_processed_id, state_lock, redis_manager

        if life_state == 0:
            logger.warning("任务已在运行中")
            return runtimeIntent_pb2.Result(msg='任务已在运行中', stateCode=0)

        # 重置消费者组位置
        try:
            redis_manager.reset_consumer_group('0-0')  # 从头开始
            logger.info("消费者组位置已重置到开头")
        except Exception as e:
            logger.error(f"重置消费者组失败: {str(e)}")
            # 即使重置失败，仍然启动服务

        # 启动处理线程
        life_state = 0
        processing_thread = threading.Thread(target=process_images, daemon=True)
        processing_thread.start()

        logger.info('-' * 10 + self.__class__.__name__ + '--start()--end' + '-' * 10)
        return runtimeIntent_pb2.Result(msg='start finished', stateCode=1)

    def store(self, request, context):
        """保存状态时只保存消息ID"""
        global current_message_id, last_processed_id

        state = {
            "last_message_id": current_message_id,
            "last_processed_id": last_processed_id
        }
        logger.info(f"保存状态: {state}")
        return runtimeIntent_pb2.Result(msg='store finished', stateCode=1, data=json.dumps(state))

    def restore(self, request, context):
        """恢复时使用保存的消息ID作为起始点"""
        global life_state, processing_thread, current_message_id, last_processed_id, state_lock, redis_manager

        logger.info('-' * 10 + self.__class__.__name__ + '--restore()--start' + '-' * 10)

        # 解析状态数据
        if request.data:
            try:
                # 提取状态数据字符串
                state_str = None
                for data_item in request.data:
                    if data_item.name == "restoreData":
                        state_str = data_item.data
                        break

                if not state_str:
                    logger.error("恢复状态失败: 未找到状态数据")
                    return runtimeIntent_pb2.Result(msg='restore failed: no state data', stateCode=0)

                # 解析状态数据
                state = json.loads(state_str)
                last_message_id = state.get("last_message_id")
                last_processed_id_val = state.get("last_processed_id")

                # 关键修改：设置消费者组位置到指定消息ID
                if last_message_id:
                    try:
                        # 设置消费者组位置到指定消息ID
                        redis_manager.reset_consumer_group(last_message_id)  #这是关键
                        logger.info(f"设置消费者组位置到消息ID: {last_message_id}")
                    except Exception as e:
                        logger.error(f"设置消费者组位置失败: {str(e)}")
                        # 回退到Stream开头
                        redis_manager.reset_consumer_group('0-0')
                        logger.warning("回退到Stream开头")

                # 设置状态
                with state_lock:
                    current_message_id = last_message_id
                    last_processed_id = last_processed_id_val
                logger.info(f"恢复状态: {state}")

                # 启动处理线程
                life_state = 0
                if not processing_thread or not processing_thread.is_alive():
                    processing_thread = threading.Thread(target=process_images, daemon=True)
                    processing_thread.start()
                    logger.info("从恢复状态启动处理线程")
            except Exception as e:
                logger.error(f"状态恢复失败: {str(e)}")
                return runtimeIntent_pb2.Result(msg='restore failed', stateCode=0)

        logger.info('-' * 10 + self.__class__.__name__ + '--restore()--end' + '-' * 10)
        return runtimeIntent_pb2.Result(msg='restore finished', stateCode=1)

    def stop(self, request, context):
        logger.info('-' * 10 + self.__class__.__name__ + '--stop()--start' + '-' * 10)
        global life_state

        life_state = -1
        if processing_thread and processing_thread.is_alive():
            processing_thread.join(timeout=5.0)

        logger.info('-' * 10 + self.__class__.__name__ + '--stop()--end' + '-' * 10)
        return runtimeIntent_pb2.Result(msg='stop finished', stateCode=1)


def process_images():
    """异步处理图片的主循环"""
    global current_message_id, last_processed_id, life_state, redis_manager, session, model_inputs, input_width, input_height, state_lock

    # 使用唯一消费者ID（时间戳+进程ID）
    consumer_id = f"yolo-consumer-{int(time.time())}-{os.getpid()}"
    logger.info(f"消费者ID: {consumer_id}")

    while life_state == 0:  # 仅在运行状态下处理
        try:
            # logger.info("尝试获取下一张图片...")
            # 获取下一张图片（从指定ID开始）
            message_id, image_id, image = redis_manager.get_next_image(consumer_id)

            if not message_id:
                # logger.info("未获取到新消息，等待100ms")
                time.sleep(10)
                continue

            # 更新状态（加锁确保一致性）
            with state_lock:
                current_message_id = message_id
                last_processed_id = image_id
            logger.info(f"开始处理图片: {image_id} (消息ID: {message_id})")

            # 处理图片
            # logger.info("开始处理图片...")
            result_image = detect_object(image, session, model_inputs, input_width, input_height)
            # logger.info("图片处理完成")

            # 保存结果
            output_dir = os.path.join(os.path.dirname(__file__), "output")
            os.makedirs(output_dir, exist_ok=True)
            output_path = os.path.join(output_dir, f"{image_id}.jpg")
            cv2.imwrite(output_path, result_image)
            # logger.info(f"结果保存到: {output_path}")

            # 确认消息
            redis_manager.ack_message(message_id)
            # logger.info(f"处理完成: {image_id}")

        except Exception as e:
            logger.error(f"处理失败: {str(e)}")

    logger.info("异步推理任务停止")


def serve(port=5123):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10), options=[
        ('grpc.max_send_message_length', 100 * 1024 * 1024),
        ('grpc.max_receive_message_length', 100 * 1024 * 1024)])
    runtimeIntent_pb2_grpc.add_RuntimeIntentServicer_to_server(RuntimeService("yolo-task"), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"gRPC服务启动，监听端口 {port}")
    server.wait_for_termination()


def gRPC_server_start():
    # port = 5123
    port = os.environ.get("PORT_FOR_RPC","5123")
    logging.info(f"RPC Port: {port}")
    if port:
        port = int(port)
        grpc_thread = threading.Thread(target=serve, daemon=True, args=(port,))
        grpc_thread.start()



if __name__ == '__main__':
    # test_redis_read()
    # 初始化模型
    model_path = os.path.join(os.path.dirname(os.path.realpath(__file__)), 'model/yolov8n.onnx')
    session, model_inputs, input_width, input_height = init_detect_model(model_path)
    logger.info(f"模型加载成功: {model_path}")

    # 初始化Redis管理器
    redis_manager = RedisStreamManager(group_name='inference_group')

    # 启动gRPC服务
    gRPC_server_start()

    logger.info('-' * 10 + 'yolo-runner' + '-' * 10)
    logger.info("等待控制命令...")

    # 主循环等待控制命令
    while True:
        time.sleep(10)

