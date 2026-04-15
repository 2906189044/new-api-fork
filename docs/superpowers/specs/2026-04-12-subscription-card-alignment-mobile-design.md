# Subscription Card Alignment And Mobile Consistency Design

## Goal

修复订阅套餐卡片在内容数量不一致时按钮不对齐的问题，并保证手机端展示与桌面端使用相同的数据内容，只调整布局列数，不丢失展示内容。

## Scope

- 保持现有套餐数据结构不变
- 不新增后台配置字段
- 仅调整订阅卡片渲染与样式约束
- 保持桌面端 `2~3` 列、手机端单列

## Root Cause

### 1. 按钮不对齐

当前卡片虽然使用了 `flex flex-col` 和 `mt-auto`，但按钮前面的内容区域高度依然取决于卖点数量。卖点少的卡片会更早结束内容流，导致按钮比其他卡片更早上移。

### 2. 手机端内容不完整

当前标题、副标题、补充文案都使用了固定行数截断和最小高度策略。这种策略在桌面端有利于保持整齐，但在手机单列场景下会裁掉用户已经配置的内容，表现为“内容不完整”。

## Design

### Card Height Strategy

- 套餐卖点统一渲染为固定槽位数
- 使用现有卖点数组，不足的部分补空白占位行
- 占位行保留高度，不显示图标和文字
- 按钮继续放在卡片底部，确保所有卡片按钮基线一致

### Responsive Content Strategy

- 桌面端保留有限截断，维持排版整齐
- 手机端去掉副标题和补充文案的行数截断
- 手机端继续使用与桌面端相同的标题、卖点、权益、补充文案数据
- 只改变列数，不改变内容来源和信息层级

## Files

- `web/src/components/topup/SubscriptionPlansCard.jsx`
- `web/src/helpers/subscriptionFormat.js`
- `web/src/helpers/subscriptionFormat.test.js`

## Testing

- 为固定槽位补齐逻辑补测试
- 为移动端是否应截断提供可测试的纯函数
- 运行前端相关单测与变更文件 ESLint
