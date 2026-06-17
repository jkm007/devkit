package service

import (
	"backend-server/internal/model"
	"backend-server/internal/repository"
	"backend-server/pkg/database"
)

// UserHomeService 用户首页服务
type UserHomeService struct {
	userRepo *repository.UserRepo
}

func NewUserHomeService() *UserHomeService {
	db := database.GetMySQL()
	return &UserHomeService{
		userRepo: repository.NewUserRepo(db),
	}
}

// HomeData 用户首页数据（移动端用）
type HomeData struct {
	Stats      HomeStats              `json:"stats"`
	Recommended []RecommendedQuestion `json:"recommended"`
}

// HomeStats 首页统计
type HomeStats struct {
	TotalQuestions    int64   `json:"totalQuestions"`
	TodayPracticeCount int    `json:"todayPracticeCount"`
	TodayCorrectRate  float64 `json:"todayCorrectRate"`
	ContinuousDays    int     `json:"continuousDays"`
}

// UserStats 用户学习统计（我的页面用）
type UserStats struct {
	TotalAnswered   int64   `json:"totalAnswered"`
	TotalCorrect    int64   `json:"totalCorrect"`
	CorrectRate     float64 `json:"correctRate"`
	ContinuousDays  int     `json:"continuousDays"`
	FavoritesCount  int64   `json:"favoritesCount"`
	WrongCount      int64   `json:"wrongCount"`
	PracticeDays    int     `json:"practiceDays"`
}

// RecommendedQuestion 推荐题目
type RecommendedQuestion struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	QuestionType string `json:"questionType"`
	Difficulty   int    `json:"difficulty"`
	CategoryName string `json:"categoryName"`
}

// GetHomeData 获取用户首页数据
func (s *UserHomeService) GetHomeData(userID uint) (*HomeData, error) {
	db := database.GetMySQL()
	data := &HomeData{}

	// 1. 总题量（已发布的题目）
	var totalQuestions int64
	db.Model(&model.Question{}).Where("status = ?", "published").Count(&totalQuestions)
	data.Stats.TotalQuestions = totalQuestions

	// 2. 今日练习次数
	var todayPracticeCount int64
	db.Model(&model.PracticeRecord{}).
		Where("user_id = ? AND DATE(created_at) = CURDATE()", userID).
		Count(&todayPracticeCount)
	data.Stats.TodayPracticeCount = int(todayPracticeCount)

	// 3. 今日正确率
	var todayTotal, todayCorrect int64
	db.Model(&model.PracticeRecord{}).
		Where("user_id = ? AND DATE(created_at) = CURDATE()", userID).
		Select("COALESCE(SUM(total), 0), COALESCE(SUM(correct), 0)").
		Row().Scan(&todayTotal, &todayCorrect)
	if todayTotal > 0 {
		data.Stats.TodayCorrectRate = float64(todayCorrect) / float64(todayTotal)
	}

	// 4. 连续打卡天数（简化逻辑：最近有练习记录的天数）
	var continuousDays int
	rows, _ := db.Raw(`
		SELECT COUNT(DISTINCT DATE(created_at)) as days
		FROM practice_records
		WHERE user_id = ?
		AND created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
	`, userID).Rows()
	if rows.Next() {
		rows.Scan(&continuousDays)
		rows.Close()
	}
	data.Stats.ContinuousDays = continuousDays

	// 5. 推荐题目（随机取3道已发布题目）
	var questions []model.Question
	db.Where("status = ?", "published").
		Order("RAND()").
		Limit(3).
		Find(&questions)

	for _, q := range questions {
		// 获取分类名
		var categoryName string
		if q.CategoryID > 0 {
			db.Table("qb_categories").Where("id = ?", q.CategoryID).Select("name").Scan(&categoryName)
		}
		data.Recommended = append(data.Recommended, RecommendedQuestion{
			ID:           q.ID,
			Title:        q.Title,
			QuestionType: q.QuestionType,
			Difficulty:   q.Difficulty,
			CategoryName: categoryName,
		})
	}

	return data, nil
}

// GetUserStats 获取用户学习统计
func (s *UserHomeService) GetUserStats(userID uint) (*UserStats, error) {
	db := database.GetMySQL()
	stats := &UserStats{}

	// 1. 总答题数和正确数
	var totalAnswered, totalCorrect int64
	db.Model(&model.PracticeRecord{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(answered), 0), COALESCE(SUM(correct), 0)").
		Row().Scan(&totalAnswered, &totalCorrect)
	stats.TotalAnswered = totalAnswered
	stats.TotalCorrect = totalCorrect
	if totalAnswered > 0 {
		stats.CorrectRate = float64(totalCorrect) / float64(totalAnswered)
	}

	// 2. 连续打卡天数
	rows, _ := db.Raw(`
		SELECT COUNT(DISTINCT DATE(created_at)) as days
		FROM practice_records
		WHERE user_id = ?
		AND created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
	`, userID).Rows()
	if rows.Next() {
		rows.Scan(&stats.ContinuousDays)
		rows.Close()
	}

	// 3. 收藏数
	db.Model(&model.UserFavorite{}).Where("user_id = ?", userID).Count(&stats.FavoritesCount)

	// 4. 错题数
	db.Model(&model.WrongBook{}).Where("user_id = ? AND is_mastered = ?", userID, false).Count(&stats.WrongCount)

	// 5. 练习天数
	rows2, _ := db.Raw(`
		SELECT COUNT(DISTINCT DATE(created_at)) as days
		FROM practice_records
		WHERE user_id = ?
	`, userID).Rows()
	if rows2.Next() {
		rows2.Scan(&stats.PracticeDays)
		rows2.Close()
	}

	return stats, nil
}
