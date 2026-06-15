#!/usr/bin/env python3
"""生成uni-app tabBar需要的PNG图标"""
from PIL import Image, ImageDraw, ImageFont
import os

OUTPUT_DIR = os.path.join(os.path.dirname(__file__), '..', 'src', 'static', 'images', 'tabbar')
os.makedirs(OUTPUT_DIR, exist_ok=True)

SIZE = 48  # 图标尺寸

def draw_home(color, filename):
    """首页图标：房子"""
    img = Image.new('RGBA', (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    # 屋顶三角形
    draw.polygon([(24, 6), (6, 24), (42, 24)], fill=color)
    # 房身
    draw.rectangle([12, 24, 36, 42], fill=color)
    # 门
    draw.rectangle([20, 30, 28, 42], fill=(255, 255, 255, 255) if color != (255, 255, 255) else (0, 0, 0, 255))
    img.save(os.path.join(OUTPUT_DIR, filename))

def draw_question(color, filename):
    """题库图标：书本"""
    img = Image.new('RGBA', (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    # 书本外框
    draw.rectangle([8, 8, 40, 40], fill=color, outline=color, width=2)
    # 书脊
    draw.line([(24, 8), (24, 40)], fill=(255, 255, 255, 255) if color != (255, 255, 255) else (0, 0, 0, 255), width=2)
    # 问号
    cx, cy = 16, 22
    draw.arc([cx-4, cy-8, cx+4, cy], 0, 270, fill=(255, 255, 255, 255) if color != (255, 255, 255) else (0, 0, 0, 255), width=2)
    draw.ellipse([cx-1, cy+4, cx+1, cy+6], fill=(255, 255, 255, 255) if color != (255, 255, 255) else (0, 0, 0, 255))
    img.save(os.path.join(OUTPUT_DIR, filename))

def draw_practice(color, filename):
    """练习图标：铅笔"""
    img = Image.new('RGBA', (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    # 笔身
    draw.polygon([(10, 38), (14, 34), (34, 14), (38, 10), (34, 14), (30, 18)], fill=color)
    # 笔尖
    draw.polygon([(8, 42), (10, 38), (14, 38)], fill=color)
    # 橡皮擦
    draw.rectangle([32, 8, 40, 16], fill=color)
    img.save(os.path.join(OUTPUT_DIR, filename))

def draw_my(color, filename):
    """我的图标：人像"""
    img = Image.new('RGBA', (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)
    # 头部
    draw.ellipse([16, 6, 32, 22], fill=color)
    # 身体
    draw.ellipse([8, 24, 40, 44], fill=color)
    img.save(os.path.join(OUTPUT_DIR, filename))

if __name__ == '__main__':
    normal_color = (153, 153, 153)  # #999999
    active_color = (24, 144, 255)   # #1890ff

    print(f"生成图标到: {OUTPUT_DIR}")

    draw_home(normal_color, 'home.png')
    draw_home(active_color, 'home-active.png')
    draw_question(normal_color, 'question.png')
    draw_question(active_color, 'question-active.png')
    draw_practice(normal_color, 'practice.png')
    draw_practice(active_color, 'practice-active.png')
    draw_my(normal_color, 'my.png')
    draw_my(active_color, 'my-active.png')

    print("✅ 图标生成完成！")
    for f in os.listdir(OUTPUT_DIR):
        if f.endswith('.png'):
            print(f"  - {f}")
