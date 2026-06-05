package captcha

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"sort"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/wenlng/go-captcha-assets/resources/fonts/fzshengsksjw"
	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/click"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"
)

// 中文字符集
var chineseChars = []string{
	"风", "限", "数", "统", "系", "配", "送", "质", "控", "流",
	"程", "工", "审", "批", "管", "理", "设", "置", "用", "户",
	"权", "限", "角", "色", "模", "块", "菜", "单", "数", "据",
	"查", "询", "添", "加", "删", "除", "修", "改", "导", "入",
	"导", "出", "上", "传", "下", "载", "打", "印", "预", "览",
	"确", "认", "取", "消", "保", "存", "提", "交", "返", "回",
	"退", "出", "登", "录", "注", "销", "密", "码", "验", "证",
}

// ==================== Slide 滑块验证码 ====================

// generateGoSlide 生成滑块验证码（动态配置）
func generateGoSlide(config *CaptchaConfig) (*CaptchaData, error) {
	width := config.SliderWidth
	height := config.SliderHeight

	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: width, Height: height}),
		slide.WithRangeGraphSize(option.RangeVal{Min: 45, Max: 55}),
	)

	// 加载内置背景图
	bgImages, err := images.GetImages()
	if err != nil || len(bgImages) == 0 {
		bgImages = []image.Image{generateDefaultBg(width, height)}
	}

	// 加载内置拼图块素材
	tileGraphs, err := tiles.GetTiles()
	if err != nil || len(tileGraphs) == 0 {
		builder.SetResources(
			slide.WithBackgrounds(bgImages),
		)
	} else {
		graphs := make([]*slide.GraphImage, 0, len(tileGraphs))
		for _, t := range tileGraphs {
			graphs = append(graphs, &slide.GraphImage{
				OverlayImage: t.OverlayImage,
				ShadowImage:  t.ShadowImage,
				MaskImage:    t.MaskImage,
			})
		}
		builder.SetResources(
			slide.WithBackgrounds(bgImages),
			slide.WithGraphImages(graphs),
		)
	}

	slideCapt := builder.Make()

	captData, err := slideCapt.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate slide captcha: %w", err)
	}

	blockData := captData.GetData()
	if blockData == nil {
		return nil, fmt.Errorf("slide captcha data is nil")
	}

	mBase64, err := captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("slide master image: %w", err)
	}
	tBase64, err := captData.GetTileImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("slide tile image: %w", err)
	}

	captchaID, _ := generateID()
	answer := GoSlideAnswer{X: int(blockData.X), Y: int(blockData.Y)}
	answerJSON, _ := json.Marshal(answer)
	if err := saveToRedis(captchaID, "slider", string(answerJSON)); err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     mBase64,
		Thumb:     tBase64,
		ThumbY:    int(blockData.Y),
		Type:      "slider",
		Width:     width,
		Height:    height,
	}, nil
}

// ==================== Puzzle 拼图验证码 ====================

// generateGoPuzzle 生成拼图验证码（动态配置）
func generateGoPuzzle(config *CaptchaConfig) (*CaptchaData, error) {
	width := config.PuzzleWidth
	height := config.PuzzleHeight

	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: width, Height: height}),
		slide.WithRangeGraphSize(option.RangeVal{Min: 50, Max: 60}),
		slide.WithEnableGraphVerticalRandom(config.PuzzleVerticalRandom),
	)

	bgImages, err := images.GetImages()
	if err != nil || len(bgImages) == 0 {
		bgImages = []image.Image{generateDefaultBg(width, height)}
	}

	tileGraphs, err := tiles.GetTiles()
	if err != nil || len(tileGraphs) == 0 {
		builder.SetResources(
			slide.WithBackgrounds(bgImages),
		)
	} else {
		graphs := make([]*slide.GraphImage, 0, len(tileGraphs))
		for _, t := range tileGraphs {
			graphs = append(graphs, &slide.GraphImage{
				OverlayImage: t.OverlayImage,
				ShadowImage:  t.ShadowImage,
				MaskImage:    t.MaskImage,
			})
		}
		builder.SetResources(
			slide.WithBackgrounds(bgImages),
			slide.WithGraphImages(graphs),
		)
	}

	dragCapt := builder.MakeDragDrop()

	captData, err := dragCapt.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate puzzle captcha: %w", err)
	}

	blockData := captData.GetData()
	if blockData == nil {
		return nil, fmt.Errorf("puzzle captcha data is nil")
	}

	mBase64, err := captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("puzzle master image: %w", err)
	}
	tBase64, err := captData.GetTileImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("puzzle tile image: %w", err)
	}

	captchaID, _ := generateID()
	answer := GoSlideAnswer{X: int(blockData.X), Y: int(blockData.Y)}
	answerJSON, _ := json.Marshal(answer)
	if err := saveToRedis(captchaID, "puzzle", string(answerJSON)); err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     mBase64,
		Thumb:     tBase64,
		ThumbY:    int(blockData.Y),
		Type:      "puzzle",
		Width:     width,
		Height:    height,
	}, nil
}

// ==================== Rotation 旋转验证码 ====================

// generateGoRotation 生成旋转验证码（动态配置）
func generateGoRotation(config *CaptchaConfig) (*CaptchaData, error) {
	size := config.RotationSize
	thumbSize := config.RotationThumbSize

	builder := rotate.NewBuilder(
		rotate.WithImageSquareSize(size),
		rotate.WithRangeAnglePos([]option.RangeVal{{Min: 30, Max: 330}}),
		rotate.WithRangeThumbImageSquareSize([]int{thumbSize, thumbSize + 20}),
		rotate.WithThumbImageAlpha(0.9),
	)

	bgImages, err := images.GetImages()
	if err != nil || len(bgImages) == 0 {
		bgImages = []image.Image{generateDefaultBg(size, size)}
	}

	builder.SetResources(
		rotate.WithImages(bgImages),
	)

	rotateCapt := builder.Make()

	captData, err := rotateCapt.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate rotate captcha: %w", err)
	}

	blockData := captData.GetData()
	if blockData == nil {
		return nil, fmt.Errorf("rotate captcha data is nil")
	}

	mBase64, err := captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("rotate master image: %w", err)
	}
	tBase64, err := captData.GetThumbImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("rotate thumb image: %w", err)
	}

	captchaID, _ := generateID()
	answer := RotationAnswer{Angle: float64(blockData.Angle)}
	answerJSON, _ := json.Marshal(answer)
	if err := saveToRedis(captchaID, "rotation", string(answerJSON)); err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     mBase64,
		Thumb:     tBase64,
		Type:      "rotation",
		Width:     size,
		Height:    size,
	}, nil
}

// ==================== Point 点选验证码 ====================

// generateGoClick 生成点选验证码（动态配置）
func generateGoClick(config *CaptchaConfig) (*CaptchaData, error) {
	width := config.PointWidth
	height := config.PointHeight
	count := config.PointCount

	builder := click.NewBuilder(
		click.WithRangeLen(option.RangeVal{Min: count, Max: count + 2}),
		click.WithRangeVerifyLen(option.RangeVal{Min: count, Max: count}),
		click.WithImageSize(option.Size{Width: width, Height: height}),
		click.WithRangeColors([]string{"#1e88e5", "#e91e63", "#ff9800", "#4caf50", "#9c27b0"}),
		click.WithDisplayShadow(true),
	)

	bgImages, err := images.GetImages()
	if err != nil || len(bgImages) == 0 {
		bgImages = []image.Image{generateDefaultBg(width, height)}
	}

	font, err := fzshengsksjw.GetFont()
	if err != nil {
		log.Printf("[Captcha] 加载字体失败: %v", err)
	}

	builder.SetResources(
		click.WithChars(chineseChars),
		click.WithFonts([]*truetype.Font{font}),
		click.WithBackgrounds(bgImages),
	)

	clickCapt := builder.Make()

	captData, err := clickCapt.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate click captcha: %w", err)
	}

	dotData := captData.GetData()
	if dotData == nil {
		return nil, fmt.Errorf("click captcha data is nil")
	}

	mBase64, err := captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("click master image: %w", err)
	}
	tBase64, err := captData.GetThumbImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("click thumb image: %w", err)
	}

	// 构建验证数据和提示文字
	points := make([]Point, 0, len(dotData))
	chars := make([]string, 0, len(dotData))
	for _, dot := range dotData {
		points = append(points, Point{X: int(dot.X), Y: int(dot.Y)})
		chars = append(chars, dot.Text)
	}

	hintText := "请依次点击【" + strings.Join(chars, "，") + "】"

	captchaID, _ := generateID()
	answer := PointAnswer{Points: points, Chars: chars}
	answerJSON, _ := json.Marshal(answer)
	if err := saveToRedis(captchaID, "point", string(answerJSON)); err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     mBase64,
		Thumb:     tBase64,
		Type:      "point",
		HintText:  hintText,
		Chars:     chars,
		Width:     width,
		Height:    height,
	}, nil
}

// ==================== Numeric 数字验证码 ====================

// generateGoNumeric 生成数字验证码（动态配置）
func generateGoNumeric(config *CaptchaConfig) (*CaptchaData, error) {
	length := config.NumericLength
	width := config.NumericWidth
	height := config.NumericHeight

	builder := click.NewBuilder(
		click.WithRangeLen(option.RangeVal{Min: length, Max: length}),
		click.WithRangeVerifyLen(option.RangeVal{Min: length, Max: length}),
		click.WithImageSize(option.Size{Width: width, Height: height}),
		click.WithRangeColors([]string{"#1e88e5", "#e91e63", "#ff9800", "#4caf50"}),
	)

	bgImages, err := images.GetImages()
	if err != nil || len(bgImages) == 0 {
		bgImages = []image.Image{generateDefaultBg(width, height)}
	}

	font, err := fzshengsksjw.GetFont()
	if err != nil {
		log.Printf("[Captcha] 加载字体失败: %v", err)
	}

	numChars := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	builder.SetResources(
		click.WithChars(numChars),
		click.WithFonts([]*truetype.Font{font}),
		click.WithBackgrounds(bgImages),
	)

	numCapt := builder.Make()

	captData, err := numCapt.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate numeric captcha: %w", err)
	}

	dotData := captData.GetData()
	if dotData == nil {
		return nil, fmt.Errorf("numeric captcha data is nil")
	}

	mBase64, err := captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, fmt.Errorf("numeric master image: %w", err)
	}

	// 按 X 坐标排序
	sortedDots := make([]*click.Dot, len(dotData))
	for i, dot := range dotData {
		sortedDots[i] = dot
	}
	sort.Slice(sortedDots, func(i, j int) bool {
		return sortedDots[i].X < sortedDots[j].X
	})

	code := ""
	for _, dot := range sortedDots {
		code += dot.Text
	}

	captchaID, _ := generateID()
	if err := saveToRedis(captchaID, "numeric", code); err != nil {
		return nil, err
	}

	return &CaptchaData{
		CaptchaID: captchaID,
		Image:     mBase64,
		Type:      "numeric",
		Width:     width,
		Height:    height,
		Length:    length,
	}, nil
}

// generateDefaultBg 生成默认背景（fallback）
func generateDefaultBg(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, image.White)
		}
	}
	return img
}

// ==================== 验证函数 ====================

// Point 点击坐标
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// GoSlideAnswer 滑块/拼图答案
type GoSlideAnswer struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// RotationAnswer 旋转答案
type RotationAnswer struct {
	Angle float64 `json:"angle"`
}

// PointAnswer 点选答案
type PointAnswer struct {
	Points []Point `json:"points"`
	Chars  []string `json:"chars"`
}

// verifySliderAnswer 验证滑块答案（使用配置容差）
func verifySliderAnswer(answerStr, value string, tolerance int) (bool, string) {
	var answer GoSlideAnswer
	if err := json.Unmarshal([]byte(decryptAnswer(answerStr)), &answer); err != nil {
		return false, "验证码数据异常"
	}

	var userAnswer struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := json.Unmarshal([]byte(value), &userAnswer); err != nil {
		return false, "提交数据格式错误"
	}

	// 滑块验证码：只检查 X 坐标，Y 使用答案值
	ok := slide.Validate(userAnswer.X, answer.Y, answer.X, answer.Y, tolerance)
	if ok {
		return true, "验证通过"
	}
	return false, fmt.Sprintf("验证失败，请重试 (X差:%d, 容差:%d)", abs(userAnswer.X-answer.X), tolerance)
}

// verifyPuzzleAnswer 验证拼图答案（使用配置容差）
func verifyPuzzleAnswer(answerStr, value string, tolerance int) (bool, string) {
	var answer GoSlideAnswer
	if err := json.Unmarshal([]byte(decryptAnswer(answerStr)), &answer); err != nil {
		return false, "验证码数据异常"
	}

	var userAnswer struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := json.Unmarshal([]byte(value), &userAnswer); err != nil {
		return false, "提交数据格式错误"
	}

	// 拼图验证码：检查 X 和 Y
	ok := slide.Validate(userAnswer.X, userAnswer.Y, answer.X, answer.Y, tolerance)
	if ok {
		return true, "验证通过"
	}
	return false, fmt.Sprintf("验证失败，请重试 (X差:%d, Y差:%d)", abs(userAnswer.X-answer.X), abs(userAnswer.Y-answer.Y))
}

// verifyRotationAnswer 验证旋转答案（使用配置容差）
func verifyRotationAnswer(answerStr, value string, tolerance int) (bool, string) {
	var answer RotationAnswer
	if err := json.Unmarshal([]byte(decryptAnswer(answerStr)), &answer); err != nil {
		return false, "验证码数据异常"
	}

	var userAnswer struct {
		Angle float64 `json:"angle"`
	}
	if err := json.Unmarshal([]byte(value), &userAnswer); err != nil {
		return false, "提交数据格式错误"
	}

	angleDiff := abs(int(userAnswer.Angle) - int(answer.Angle))

	ok := rotate.Validate(int(userAnswer.Angle), int(answer.Angle), tolerance)
	if ok {
		return true, "验证通过"
	}
	return false, fmt.Sprintf("角度偏差过大，请重试 (差值:%d°, 容差:%d°)", angleDiff, tolerance)
}

// verifyPointAnswer 验证点选答案（使用配置容差）
func verifyPointAnswer(answerStr string, points []Point, tolerance int) (bool, string) {
	var answer PointAnswer
	if err := json.Unmarshal([]byte(decryptAnswer(answerStr)), &answer); err != nil {
		return false, "验证码数据异常"
	}

	if len(points) != len(answer.Points) {
		return false, fmt.Sprintf("验证点数不匹配 (提交:%d, 期望:%d)", len(points), len(answer.Points))
	}

	// 依次验证每个点击点
	for i, p := range points {
		if i >= len(answer.Points) {
			break
		}
		ok := click.Validate(p.X, p.Y, answer.Points[i].X, answer.Points[i].Y, 60, 60, tolerance)
		if !ok {
			return false, fmt.Sprintf("点击位置不正确 (第%d个点)", i+1)
		}
	}
	return true, "验证通过"
}