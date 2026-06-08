package service

import (
	"claude-manager/internal/model"
	"claude-manager/internal/repository"
	"claude-manager/pkg/storage"
	"fmt"
)

// RoutingService 路由规则服务
type RoutingService struct {
	routingRepo *repository.TagRoutingRepo
}

// NewRoutingService 创建路由规则服务
func NewRoutingService(routingRepo *repository.TagRoutingRepo) *RoutingService {
	return &RoutingService{
		routingRepo: routingRepo,
	}
}

// GetAllRules 获取所有路由规则
func (s *RoutingService) GetAllRules() ([]model.TagRouting, error) {
	return s.routingRepo.GetAll()
}

// GetEnabledRules 获取启用的路由规则
func (s *RoutingService) GetEnabledRules() ([]model.TagRouting, error) {
	return s.routingRepo.GetEnabled()
}

// GetRuleByID 根据ID获取规则
func (s *RoutingService) GetRuleByID(id int64) (*model.TagRouting, error) {
	return s.routingRepo.GetByID(id)
}

// CreateRule 创建路由规则
func (s *RoutingService) CreateRule(rule *model.TagRouting) error {
	// 如果设置为默认规则，取消其他默认规则
	if rule.IsDefault {
		if err := s.clearDefaultRule(); err != nil {
			return err
		}
	}
	return s.routingRepo.Create(rule)
}

// UpdateRule 更新路由规则
func (s *RoutingService) UpdateRule(rule *model.TagRouting) error {
	// 如果设置为默认规则，取消其他默认规则
	if rule.IsDefault {
		if err := s.clearDefaultRule(); err != nil {
			return err
		}
	}

	if err := s.routingRepo.Update(rule); err != nil {
		return err
	}

	// 刷新路由引擎
	return storage.RefreshRoutingEngine()
}

// DeleteRule 删除路由规则
func (s *RoutingService) DeleteRule(id int64) error {
	rule, err := s.routingRepo.GetByID(id)
	if err != nil {
		return err
	}
	if rule.IsDefault {
		return fmt.Errorf("默认规则不允许删除")
	}

	if err := s.routingRepo.Delete(id); err != nil {
		return err
	}

	// 刷新路由引擎
	return storage.RefreshRoutingEngine()
}

// UpdateStatus 更新规则状态
func (s *RoutingService) UpdateStatus(id int64, status int8) error {
	if err := s.routingRepo.UpdateStatus(id, status); err != nil {
		return err
	}
	return storage.RefreshRoutingEngine()
}

// UpdatePriority 更新规则优先级
func (s *RoutingService) UpdatePriority(id int64, priority int) error {
	if err := s.routingRepo.UpdatePriority(id, priority); err != nil {
		return err
	}
	return storage.RefreshRoutingEngine()
}

// BatchUpdatePriority 批量更新优先级
func (s *RoutingService) BatchUpdatePriority(priorities map[int64]int) error {
	if err := s.routingRepo.BatchUpdatePriority(priorities); err != nil {
		return err
	}
	return storage.RefreshRoutingEngine()
}

// TestRule 测试规则匹配
func (s *RoutingService) TestRule(ruleID int64, tags []storage.RoutingTag) (bool, error) {
	engine := storage.GetRoutingEngine()
	if engine == nil {
		return false, fmt.Errorf("路由引擎未初始化")
	}
	return engine.TestRule(ruleID, tags)
}

// TestRoute 测试文件路由
func (s *RoutingService) TestRoute(filename, contentType, source string) (*storage.RoutingResult, []storage.RoutingTag, error) {
	return storage.Route(filename, contentType, source)
}

// clearDefaultRule 清除其他默认规则
func (s *RoutingService) clearDefaultRule() error {
	rules, err := s.routingRepo.GetAll()
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.IsDefault {
			rule.IsDefault = false
			if err := s.routingRepo.Update(&rule); err != nil {
				return err
			}
		}
	}
	return nil
}
