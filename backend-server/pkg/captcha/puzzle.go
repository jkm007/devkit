package captcha

import (
	"encoding/json"
	"image"
	"image/color"
)

const (
	puzzleW     = 320
	puzzleH     = 200
	puzPieceW   = 60
	puzPieceH   = 60
	puzzleTabR  = 12
)

// PuzzleAnswer 拼图答案
type PuzzleAnswer struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// generatePuzzle 生成拼图验证码
func generatePuzzle() (*CaptchaData, error) {
	bg := drawBeautifulBg(puzzleW, puzzleH)

	gapX := randInt(puzPieceW+30, puzzleW-puzPieceW*2-30)
	gapY := randInt(30, puzzleH-puzPieceH-30)

	piece := extractPuzzlePiece(bg, gapX, gapY)
	cutPuzzleSlot(bg, gapX, gapY)

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

	answer, _ := json.Marshal(PuzzleAnswer{X: gapX, Y: gapY})
	if err := saveToRedis(captchaID, "puzzle", string(answer)); err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     bgStr,
		Thumb:     pieceStr,
		Type:      "puzzle",
	}, nil
}

// verifyPuzzle 验证拼图答案，容差 ±5px
func verifyPuzzle(answerStr, value string) (bool, string) {
	var answer PuzzleAnswer
	if err := json.Unmarshal([]byte(decryptAnswer(answerStr)), &answer); err != nil {
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

// isInsidePuzzlePiece 判断点是否在拼图块形状内
// 形状：中间矩形 + 顶部圆形凸起 + 左侧凹陷 + 右侧凸起
func isInsidePuzzlePiece(px, py, w, h int) bool {
	cx := w / 2
	r := float64(puzzleTabR)

	// 中间矩形主体
	if px >= 0 && px < w && py >= puzzleTabR && py < h-puzzleTabR {
		return true
	}

	// 顶部凸起（圆形）
	dx := float64(px - cx)
	dy := float64(py - puzzleTabR)
	if dx*dx+dy*dy <= r*r {
		return true
	}

	// 左侧凹陷
	ldx := float64(px)
	ldy := float64(py - h/2)
	if ldx*ldx+ldy*ldy <= r*r {
		return false
	}

	// 右侧凸起
	rdx := float64(px - w)
	rdy := float64(py - h/2)
	if px >= w-puzzleTabR && rdx*rdx+rdy*rdy <= r*r {
		return true
	}

	return false
}

// extractPuzzlePiece 从背景图中提取拼图块（带凸起形状）
func extractPuzzlePiece(bg *image.RGBA, x, y int) *image.RGBA {
	padX := puzzleTabR + 2
	padY := puzzleTabR + 2
	canvasW := puzPieceW + padX*2
	canvasH := puzPieceH + padY*2
	piece := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	for py := 0; py < canvasH; py++ {
		for px := 0; px < canvasW; px++ {
			localX := px - padX
			localY := py - padY
			if isInsidePuzzlePiece(localX, localY, puzPieceW, puzPieceH) {
				bgX := x + localX
				bgY := y + localY
				if bgX >= 0 && bgX < puzzleW && bgY >= 0 && bgY < puzzleH {
					piece.Set(px, py, bg.RGBAAt(bgX, bgY))
				} else {
					piece.Set(px, py, color.RGBA{R: 200, G: 200, B: 200, A: 255})
				}
			}
		}
	}

	drawPuzzleBorder(piece, canvasW, canvasH, padX, padY)
	return piece
}

// cutPuzzleSlot 在背景图上扣出拼图缺口
func cutPuzzleSlot(bg *image.RGBA, x, y int) {
	for py := 0; py < puzPieceH; py++ {
		for px := 0; px < puzPieceW; px++ {
			if isInsidePuzzlePiece(px, py, puzPieceW, puzPieceH) {
				sx, sy := x+px, y+py
				if sx >= 0 && sx < puzzleW && sy >= 0 && sy < puzzleH {
					orig := bg.RGBAAt(sx, sy)
					bg.Set(sx, sy, color.RGBA{
						R: uint8(int(orig.R) * 30 / 255),
						G: uint8(int(orig.G) * 30 / 255),
						B: uint8(int(orig.B) * 30 / 255),
						A: 220,
					})
				}
			}
		}
	}
}

// drawPuzzleBorder 给拼图块画白色描边
func drawPuzzleBorder(piece *image.RGBA, canvasW, canvasH, padX, padY int) {
	borderColor := color.RGBA{R: 255, G: 255, B: 255, A: 200}
	for py := 1; py < canvasH-1; py++ {
		for px := 1; px < canvasW-1; px++ {
			localX := px - padX
			localY := py - padY
			if !isInsidePuzzlePiece(localX, localY, puzPieceW, puzPieceH) {
				continue
			}
			neighbors := [][2]int{{px - 1, py}, {px + 1, py}, {px, py - 1}, {px, py + 1}}
			for _, n := range neighbors {
				nx, ny := n[0], n[1]
				if nx < 0 || nx >= canvasW || ny < 0 || ny >= canvasH {
					piece.Set(px, py, borderColor)
					break
				}
				if !isInsidePuzzlePiece(nx-padX, ny-padY, puzPieceW, puzPieceH) {
					piece.Set(px, py, borderColor)
					break
				}
			}
		}
	}
}
