package service

import (
	"strings"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

var subscriptionPlanUsableGroups = map[string][]string{
	"plan_codex_basic":    {"gpt_core"},
	"plan_codex_advanced": {"gpt_core", "gpt52_unlimited"},
	"plan_coding_all":     {"gpt_core", "gpt52_unlimited", "claude_core", "gemini_core", "upstream"},
}

func GetUserUsableGroups(userGroup string) map[string]string {
	return getUserUsableGroupsBase(userGroup)
}

func getUserUsableGroupsBase(userGroup string) map[string]string {
	groupsCopy := setting.GetUserUsableGroupsCopy()
	if setting.DefaultUseAutoGroup {
		groupsCopy["auto"] = setting.GetUsableGroupDescription("auto")
	}
	if userGroup != "" {
		specialSettings, b := ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.Get(userGroup)
		if b {
			// 处理特殊可用分组
			for specialGroup, desc := range specialSettings {
				if strings.HasPrefix(specialGroup, "-:") {
					// 移除分组
					groupToRemove := strings.TrimPrefix(specialGroup, "-:")
					delete(groupsCopy, groupToRemove)
				} else if strings.HasPrefix(specialGroup, "+:") {
					// 添加分组
					groupToAdd := strings.TrimPrefix(specialGroup, "+:")
					groupsCopy[groupToAdd] = desc
				} else {
					// 直接添加分组
					groupsCopy[specialGroup] = desc
				}
			}
		}
		// 如果userGroup不在UserUsableGroups中，返回UserUsableGroups + userGroup
		if _, ok := groupsCopy[userGroup]; !ok {
			groupsCopy[userGroup] = "用户分组"
		}
	}
	return groupsCopy
}

func bonusScopeUsableGroups(scope string) []string {
	switch strings.TrimSpace(scope) {
	case "gpt_series":
		return []string{"gpt_core", "gpt52_unlimited"}
	default:
		return []string{}
	}
}

func GetUserUsableGroupsForUser(userId int, userGroup string) map[string]string {
	groupsCopy := getUserUsableGroupsBase(userGroup)
	for _, group := range subscriptionPlanUsableGroups[strings.TrimSpace(userGroup)] {
		if _, ok := groupsCopy[group]; ok {
			continue
		}
		desc := setting.GetUsableGroupDescription(group)
		if strings.TrimSpace(desc) == "" {
			desc = "订阅套餐权益"
		}
		groupsCopy[group] = desc
	}
	if userId <= 0 {
		return groupsCopy
	}
	scopes, err := model.GetActiveStackableBonusScopes(userId)
	if err != nil {
		return groupsCopy
	}
	for _, scope := range scopes {
		for _, group := range bonusScopeUsableGroups(scope) {
			if _, ok := groupsCopy[group]; ok {
				continue
			}
			desc := setting.GetUsableGroupDescription(group)
			if strings.TrimSpace(desc) == "" {
				desc = "附加订阅权益"
			}
			groupsCopy[group] = desc
		}
	}
	return groupsCopy
}

func GroupInUserUsableGroups(userGroup, groupName string) bool {
	_, ok := GetUserUsableGroups(userGroup)[groupName]
	return ok
}

func GroupInUserUsableGroupsForUser(userId int, userGroup, groupName string) bool {
	_, ok := GetUserUsableGroupsForUser(userId, userGroup)[groupName]
	return ok
}

// GetUserAutoGroup 根据用户分组获取自动分组设置
func GetUserAutoGroup(userGroup string) []string {
	groups := GetUserUsableGroups(userGroup)
	autoGroups := make([]string, 0)
	for _, group := range setting.GetAutoGroups() {
		if _, ok := groups[group]; ok {
			autoGroups = append(autoGroups, group)
		}
	}
	return autoGroups
}

func GetUserAutoGroupForUser(userId int, userGroup string) []string {
	groups := GetUserUsableGroupsForUser(userId, userGroup)
	autoGroups := make([]string, 0)
	for _, group := range setting.GetAutoGroups() {
		if _, ok := groups[group]; ok {
			autoGroups = append(autoGroups, group)
		}
	}
	return autoGroups
}

// GetUserGroupRatio 获取用户使用某个分组的倍率
// userGroup 用户分组
// group 需要获取倍率的分组
func GetUserGroupRatio(userGroup, group string) float64 {
	ratio, ok := ratio_setting.GetGroupGroupRatio(userGroup, group)
	if ok {
		return ratio
	}
	return ratio_setting.GetGroupRatio(group)
}
