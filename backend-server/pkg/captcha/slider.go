package captcha

import (
	"encoding/json"
	"image"
	"image/color"
)

const (
	sliderWidth  = 320
	sliderHeight = 200
	pieceW       = 50
	pieceH       = 50
)

// SliderAnswer 滑块答案
type SliderAnswer struct {
	X int `json:"x"` // 缺口 X 坐标
}

// generateSlider 生成滑块验证码
// 返回带缺口的背景图 + 拼图块缩略图
func generateSlider() (*CaptchaData, error) {
	bg := drawBeautifulBg(sliderWidth, sliderHeight)

	gapX := randInt(pieceW+20, sliderWidth-pieceW*2-20)
	gapY := sliderHeight/2 - pieceH/2

	// 提取拼图块（扣洞之前）
	piece := extractPiece(bg, gapX, gapY)

	// 在背景上扣出缺口
	cutSlot(bg, gapX, gapY)

	bgStr, err := toBase64(bg)
	if err != nil {
		return nil, err
	}
	pieceStr, err := toBase64(piece)
	if err != nil {
		return nil, err
	}

	captchaID, err := generateID()
	if err != nil {
		return nil, err
	}

	answer, _ := json.Marshal(SliderAnswer{X: gapX})
	startTime, err := saveToRedis(captchaID, "slider", string(answer))
	if err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     bgStr,
		Thumb:     pieceStr,
		Type:      "slider",
		StartTime: startTime,
	}, nil
}

// verifySlider 验证滑块答案，容差 ±5px
func verifySlider(answerStr, value string) (bool, string) {
	var answer SliderAnswer
	decrypted, err := decryptAnswer(answerStr)
	if err != nil {
		return false, "验证码数据异常"
	}
	if err := json.Unmarshal([]byte(decrypted), &answer); err != nil {
		return false, "验证码数据异常"
	}
	var userAnswer struct {
		X int `json:"x"`
	}
	if err := json.Unmarshal([]byte(value), &userAnswer); err != nil {
		return false, "提交数据格式错误"
	}
	if abs(userAnswer.X-answer.X) <= 5 {
		return true, "验证通过"
	}
	return false, "验证失败，请重试"
}

// extractPiece 从背景图中提取拼图块
func extractPiece(bg *image.RGBA, x, y int) *image.RGBA {
	piece := image.NewRGBA(image.Rect(0, 0, pieceW, pieceH))
	for py := 0; py < pieceH; py++ {
		for px := 0; px < pieceW; px++ {
			sx, sy := x+px, y+py
			if sx >= 0 && sx < sliderWidth && sy >= 0 && sy < sliderHeight {
				piece.Set(px, py, bg.RGBAAt(sx, sy))
			}
		}
	}
	drawPieceBorder(piece, pieceW, pieceH)
	return piece
}

// cutSlot 在背景图上扣出缺口
func cutSlot(bg *image.RGBA, x, y int) {
	for py := 0; py < pieceH; py++ {
		for px := 0; px < pieceW; px++ {
			sx, sy := x+px, y+py
			if sx >= 0 && sx < sliderWidth && sy >= 0 && sy < sliderHeight {
				orig := bg.RGBAAt(sx, sy)
				bg.Set(sx, sy, color.RGBA{
					R: uint8(int(orig.R) * 40 / 255),
					G: uint8(int(orig.G) * 40 / 255),
					B: uint8(int(orig.B) * 40 / 255),
					A: 200,
				})
			}
		}
	}
}

// drawPieceBorder 给拼图块画白色边框
func drawPieceBorder(piece *image.RGBA, w, h int) {
	c := color.RGBA{R: 255, G: 255, B: 255, A: 220}
	for i := 0; i < w; i++ {
		piece.Set(i, 0, c)
		piece.Set(i, h-1, c)
	}
	for i := 0; i < h; i++ {
		piece.Set(0, i, c)
		piece.Set(w-1, i, c)
	}
}
