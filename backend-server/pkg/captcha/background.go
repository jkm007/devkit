package captcha

import (
	"image"
	"image/color"
	"math"
)

// 调色板集合（每组 3 个主色）
var palettes = [][3]color.RGBA{
	{{R: 58, G: 124, B: 195, A: 255}, {R: 120, G: 190, B: 230, A: 255}, {R: 200, G: 230, B: 245, A: 255}},
	{{R: 180, G: 80, B: 100, A: 255}, {R: 230, G: 150, B: 120, A: 255}, {R: 250, G: 220, B: 200, A: 255}},
	{{R: 60, G: 140, B: 90, A: 255}, {R: 120, G: 200, B: 140, A: 255}, {R: 200, G: 240, B: 210, A: 255}},
	{{R: 140, G: 90, B: 170, A: 255}, {R: 190, G: 140, B: 210, A: 255}, {R: 230, G: 200, B: 240, A: 255}},
	{{R: 200, G: 140, B: 50, A: 255}, {R: 240, G: 190, B: 100, A: 255}, {R: 255, G: 235, B: 180, A: 255}},
	{{R: 70, G: 130, B: 180, A: 255}, {R: 100, G: 170, B: 210, A: 255}, {R: 180, G: 220, B: 240, A: 255}},
	{{R: 190, G: 100, B: 70, A: 255}, {R: 230, G: 160, B: 120, A: 255}, {R: 250, G: 220, B: 190, A: 255}},
	{{R: 80, G: 160, B: 160, A: 255}, {R: 140, G: 210, B: 200, A: 255}, {R: 200, G: 240, B: 235, A: 255}},
}

// 2D 简化 Perlin 噪声
func perlinNoise(x, y float64) float64 {
	// 简化的梯度噪声
	ix := int(math.Floor(x))
	iy := int(math.Floor(y))
	fx := x - float64(ix)
	fy := y - float64(iy)

	// 平滑插值
	u := fx * fx * (3 - 2*fx)
	v := fy * fy * (3 - 2*fy)

	// 四个角的伪随机梯度点积
	n00 := pseudoRandom(ix, iy)
	n10 := pseudoRandom(ix+1, iy)
	n01 := pseudoRandom(ix, iy+1)
	n11 := pseudoRandom(ix+1, iy+1)

	// 双线性插值
	n0 := n00*(1-u) + n10*u
	n1 := n01*(1-u) + n11*u
	return n0*(1-v) + n1*v
}

// pseudoRandom 伪随机数生成（基于整数坐标）
func pseudoRandom(x, y int) float64 {
	n := x*374761393 + y*668265263
	n = (n ^ (n >> 13)) * 1274126177
	n = n ^ (n >> 16)
	return float64(n%1000) / 1000.0
}

// fbm 分形布朗运动（多层噪声叠加）
func fbm(x, y float64, octaves int) float64 {
	value := 0.0
	amplitude := 1.0
	frequency := 1.0
	maxValue := 0.0

	for i := 0; i < octaves; i++ {
		value += amplitude * perlinNoise(x*frequency, y*frequency)
		maxValue += amplitude
		amplitude *= 0.5
		frequency *= 2.0
	}

	return value / maxValue
}

// drawBeautifulBg 生成漂亮的抽象图案背景
func drawBeautifulBg(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// 随机选择调色板
	palette := palettes[randInt(0, len(palettes))]

	// 噪声参数（随机偏移避免重复）
	offX := float64(randInt(0, 1000)) / 100.0
	offY := float64(randInt(0, 1000)) / 100.0
	scale := 0.008 + float64(randInt(0, 50))/1000.0

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// 多层噪声
			nx := float64(x)*scale + offX
			ny := float64(y)*scale + offY

			n1 := fbm(nx, ny, 4)             // 基础形状
			n2 := fbm(nx*1.5+5.3, ny*1.5+2.7, 3) // 细节纹理
			n3 := fbm(nx*0.5+11.1, ny*0.5+7.9, 2) // 大尺度变化

			// 混合颜色
			t1 := clampFloat(n1)
			t2 := clampFloat(n2)
			t3 := clampFloat(n3)

			// 三色混合
			r := float64(palette[0].R)*t1 + float64(palette[1].R)*t2*0.5 + float64(palette[2].R)*t3*0.3
			g := float64(palette[0].G)*t1 + float64(palette[1].G)*t2*0.5 + float64(palette[2].G)*t3*0.3
			b := float64(palette[0].B)*t1 + float64(palette[1].B)*t2*0.5 + float64(palette[2].B)*t3*0.3

			// 添加微弱网格纹理
			gridX := math.Sin(float64(x)*0.15) * 8
			gridY := math.Sin(float64(y)*0.15) * 8
			grid := (gridX + gridY) * 0.3

			img.Set(x, y, color.RGBA{
				R: uint8(clamp(int(r + grid))),
				G: uint8(clamp(int(g + grid))),
				B: uint8(clamp(int(b + grid))),
				A: 255,
			})
		}
	}

	// 添加装饰性几何图形
	drawGeometricShapes(img, w, h)

	return img
}

// drawGeometricShapes 绘制装饰性几何图形
func drawGeometricShapes(img *image.RGBA, w, h int) {
	// 随机圆形（半透明）
	numCircles := randInt(3, 8)
	for i := 0; i < numCircles; i++ {
		cx := randInt(0, w)
		cy := randInt(0, h)
		r := randInt(20, 80)
		alpha := uint8(randInt(15, 50))
		cc := color.RGBA{
			R: uint8(randInt(100, 255)),
			G: uint8(randInt(100, 255)),
			B: uint8(randInt(100, 255)),
			A: alpha,
		}
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if dx*dx+dy*dy <= r*r {
					px, py := cx+dx, cy+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						orig := img.RGBAAt(px, py)
						t := float64(cc.A) / 255.0
						img.Set(px, py, color.RGBA{
							R: uint8(float64(orig.R)*(1-t) + float64(cc.R)*t),
							G: uint8(float64(orig.G)*(1-t) + float64(cc.G)*t),
							B: uint8(float64(orig.B)*(1-t) + float64(cc.B)*t),
							A: orig.A,
						})
					}
				}
			}
		}
	}

	// 随机矩形（半透明）
	numRects := randInt(2, 5)
	for i := 0; i < numRects; i++ {
		rx := randInt(0, w-60)
		ry := randInt(0, h-40)
		rw := randInt(30, 100)
		rh := randInt(20, 80)
		alpha := uint8(randInt(10, 35))
		rc := color.RGBA{
			R: uint8(randInt(150, 255)),
			G: uint8(randInt(150, 255)),
			B: uint8(randInt(150, 255)),
			A: alpha,
		}
		for y := ry; y < ry+rh && y < h; y++ {
			for x := rx; x < rx+rw && x < w; x++ {
				if x >= 0 && y >= 0 {
					orig := img.RGBAAt(x, y)
					t := float64(rc.A) / 255.0
					img.Set(x, y, color.RGBA{
						R: uint8(float64(orig.R)*(1-t) + float64(rc.R)*t),
						G: uint8(float64(orig.G)*(1-t) + float64(rc.G)*t),
						B: uint8(float64(orig.B)*(1-t) + float64(rc.B)*t),
						A: orig.A,
					})
				}
			}
		}
	}

	// 装饰线
	numLines := randInt(2, 6)
	for i := 0; i < numLines; i++ {
		lc := color.RGBA{
			R: uint8(randInt(150, 255)),
			G: uint8(randInt(150, 255)),
			B: uint8(randInt(150, 255)),
			A: uint8(randInt(30, 80)),
		}
		drawLine(img, randInt(0, w), randInt(0, h), randInt(0, w), randInt(0, h), lc)
	}
}

// clampFloat 将 [0,1] 范围外的值截断
func clampFloat(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
