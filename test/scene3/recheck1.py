import argparse
import json
import os
import requests

def read_and_replace(file_path, placeholder, replacement):
    """读取文件内容并替换指定占位符"""
    try:
        # 验证文件是否存在
        if not os.path.isfile(file_path):
            raise FileNotFoundError(f"文件 '{file_path}' 不存在")

        # 读取并替换内容
        with open(file_path, 'r', encoding='utf-8') as file:
            content = file.read()

        content = content.replace(f"{{{placeholder}}}", replacement)
        return content

    except (FileNotFoundError, ValueError) as e:
        print(f"错误：{e}")
        return None
    except Exception as e:
        print(f"读取文件时发生未知错误：{e}")
        return None

def send_json_request(url, ip, port, namespace, data):
    """
    发送JSON格式的HTTP请求
    :param url: 请求的基本URL
    :param ip: 服务器IP地址
    :param port: 服务器端口
    :param namespace: 命名空间
    :param data: 要发送的数据
    """
    try:
        # 构建完整的请求URL
        full_url = f"http://{ip}:{port}/{url}/v1/task?Namespace={namespace}"

        # 确保数据是JSON格式
        if not isinstance(data, str):
            data = json.dumps(data)
            
        headers = {'accept': 'application/json', 'Content-Type': 'application/json'}
        response = requests.post(full_url, data=data.encode('utf-8'), headers=headers)

        # 检查响应状态
        if response.status_code == 200:
            try:
                json_response = response.json()
                print(f"请求成功 (状态码: {response.status_code})")
                print(f"JSON响应: {json.dumps(json_response, indent=2)}")
            except json.JSONDecodeError:
                print(f"请求成功 (状态码: {response.status_code})，但响应不是JSON格式")
                print(f"响应内容: {response.text}")
        else:
            print(f"请求失败 (状态码: {response.status_code})")
            print(f"响应内容: {response.text}")
            
    except Exception as e:
        print(f"发送请求时发生错误：{e}")

# 主函数
def main():
    print('running')
    parser = argparse.ArgumentParser(description="替换文件内容并发送JSON请求")
    parser.add_argument("file_path", help="要读取的文件绝对路径")
    parser.add_argument("replacement", help="替换后的内容")
    parser.add_argument("--placeholder", default="which_device", help="要替换的占位符 (默认: which_device)")
    parser.add_argument("--ip", default="127.0.0.1", help="服务器IP地址 (默认: 127.0.0.1)")
    parser.add_argument("--port", default="8899", help="服务器端口 (默认: 8899)")
    parser.add_argument("--namespace", default="test", help="命名空间 (默认: test)")
    parser.add_argument("--url", default="framework", help="请求的基本URL (默认: framework)")

    args = parser.parse_args()

    print(f"替换内容: {args.replacement}")
    
    # 读取文件并替换内容
    replaced_content = read_and_replace(args.file_path, args.placeholder, args.replacement)
    
    if replaced_content is not None:
        print("替换后的内容:")
        print(replaced_content)
        
        # 将替换后的内容以JSON格式发送请求
        send_json_request(args.url, args.ip, args.port, args.namespace, replaced_content)
    else:
        print("无法读取或替换文件内容，程序终止")

if __name__ == '__main__':
    main()
