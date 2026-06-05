package captcha

import (
	"crypto/rand"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/big"
	"strings"
)

// 7x11 位图字体 (数字 0-9)
// 每个数字用 11 行 x 7 列的位图表示
var digitBitmaps = map[rune][]string{
	'0': {
		"  ###  ",
		" #   # ",
		"#     #",
		"#     #",
		"#     #",
		"#     #",
		"#     #",
		"#     #",
		"#     #",
		" #   # ",
		"  ###  ",
	},
	'1': {
		"   #   ",
		"  ##   ",
		" # #   ",
		"   #   ",
		"   #   ",
		"   #   ",
		"   #   ",
		"   #   ",
		"   #   ",
		"   #   ",
		" ##### ",
	},
	'2': {
		"  ###  ",
		" #   # ",
		"#     #",
		"      #",
		"     # ",
		"    #  ",
		"   #   ",
		"  #    ",
		" #     ",
		"#      ",
		"#######",
	},
	'3': {
		"  ###  ",
		" #   # ",
		"#     #",
		"      #",
		"    ## ",
		"      #",
		"      #",
		"      #",
		"#     #",
		" #   # ",
		"  ###  ",
	},
	'4': {
		"    ## ",
		"   # # ",
		"  #  # ",
		" #   # ",
		"#    # ",
		"#######",
		"     # ",
		"     # ",
		"     # ",
		"     # ",
		"     # ",
	},
	'5': {
		"#######",
		"#      ",
		"#      ",
		"#      ",
		"###### ",
		"      #",
		"      #",
		"      #",
		"#     #",
		" #   # ",
		"  ###  ",
	},
	'6': {
		"  #### ",
		" #     ",
		"#      ",
		"#      ",
		"###### ",
		"#     #",
		"#     #",
		"#     #",
		"#     #",
		" #   # ",
		"  ###  ",
	},
	'7': {
		"#######",
		"      #",
		"     # ",
		"    #  ",
		"   #   ",
		"  #    ",
		"  #    ",
		"  #    ",
		"  #    ",
		"  #    ",
		"  #    ",
	},
	'8': {
		"  ###  ",
		" #   # ",
		"#     #",
		"#     #",
		" #   # ",
		"  ###  ",
		" #   # ",
		"#     #",
		"#     #",
		" #   # ",
		"  ###  ",
	},
	'9': {
		"  ###  ",
		" #   # ",
		"#     #",
		"#     #",
		" #   # ",
		"  #### ",
		"     # ",
		"     # ",
		"     # ",
		"    #  ",
		" ####  ",
	},
}

// GenerateNumeric 生成数字验证码
func GenerateNumeric() (*CaptchaData, error) {
	code, err := generateCode(4)
	if err != nil {
		return nil, err
	}

	imgData, err := renderDigitCaptcha(code, 200, 80)
	if err != nil {
		return nil, err
	}

	captchaID, err := generateID()
	if err != nil {
		return nil, err
	}

	startTime, err := saveToRedis(captchaID, "numeric", code)
	if err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     imgData,
		Type:      "numeric",
		StartTime: startTime,
	}, nil
}

// generateCode 生成随机数字字符串
func generateCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}
	return string(code), nil
}

// renderDigitCaptcha 渲染数字验证码图片
func renderDigitCaptcha(code string, width, height int) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 背景色（浅灰）
	bgColor := color.RGBA{R: 245, G: 245, B: 245, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// 绘制干扰线（彩色）
	for i := 0; i < 6; i++ {
		c := color.RGBA{
			R: uint8(randInt(80, 200)),
			G: uint8(randInt(80, 200)),
			B: uint8(randInt(80, 200)),
			A: 200,
		}
		drawLine(img, randInt(0, width), randInt(0, height),
			randInt(0, width), randInt(0, height), c)
	}

	// 绘制干扰点
	for i := 0; i < 30; i++ {
		x := randInt(0, width)
		y := randInt(0, height)
		img.Set(x, y, color.RGBA{
			R: uint8(randInt(100, 200)),
			G: uint8(randInt(100, 200)),
			B: uint8(randInt(100, 200)),
			A: 180,
		})
	}

	// 绘制数字
	startX := 20
	charWidth := 40 // 每个字符的宽度
	for i, ch := range code {
		// 随机颜色
		textColor := color.RGBA{
			R: uint8(randInt(20, 100)),
			G: uint8(randInt(20, 100)),
			B: uint8(randInt(20, 100)),
			A: 255,
		}
		x := startX + i*charWidth
		// 随机 Y 偏移
		y := randInt(15, 35)
		drawDigit(img, ch, x, y, textColor)
	}

	// 编码为 base64
	var buf strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := png.Encode(encoder, img); err != nil {
		return "", err
	}
	encoder.Close()

	return "data:image/png;base64," + buf.String(), nil
}

// drawLine 绘制直线（Bresenham 算法）
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	for {
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
			img.Set(x1, y1, c)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// drawDigit 绘制单个数字（7x11 位图）
func drawDigit(img *image.RGBA, ch rune, startX, startY int, c color.RGBA) {
	bitmap, ok := digitBitmaps[ch]
	if !ok {
		return
	}

	scale := 3 // 放大倍数
	for row, line := range bitmap {
		for col, pixel := range line {
			if pixel == '#' {
				// 绘制放大的像素块
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						x := startX + col*scale + dx
						y := startY + row*scale + dy
						if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
							img.Set(x, y, c)
						}
					}
				}
			}
		}
	}
}
