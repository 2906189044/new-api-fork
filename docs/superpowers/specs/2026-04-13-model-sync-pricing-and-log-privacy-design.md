# Model Sync, Pricing, And Log Privacy Design

## Goal

完成四项调整，并保持现有套餐分组和自动路由结构不被打乱：

1. 从模型广场移除 `Claude Opus 4.6` 和 `Claude Sonnet 4.6` 两个旧公开模型。
2. 将上游 `CLI API` 当前新增的 GPT/Claude 模型同步到 `New API`，同时让它们能在模型广场展示并实际调用。
3. 调整模型分组倍率：`claude_core = 4`，`gemini_core = 1.5`，原有模型和 `gpt52_unlimited` 的特殊逻辑保持不变。
4. 隐藏使用日志里由模型映射产生的“实际模型”展示，不再向前端暴露真实上游模型名。

## Scope

本次会同时修改：

- `models` 模型元数据
- `channels / abilities / model_mapping`
- `ModelRatio / CompletionRatio / GroupRatio`
- 使用日志展示逻辑

本次不会修改：

- 现有套餐与订阅分组结构
- `AutoGroups` 顺序
- `gpt52_unlimited` 对 `gpt-5.2 / gpt-5.2-codex` 的 0 倍率逻辑
- 钱包、订阅、支付链路

## Current State

服务器当前状态已经确认：

- 上游 `CLI API` 已提供以下新增可用模型：
  - GPT：`gpt-5`、`gpt-5-codex`、`gpt-5-codex-mini`、`gpt-5.1`、`gpt-5.1-codex`、`gpt-5.1-codex-max`、`gpt-5.1-codex-mini`
  - Claude：`claude-opus-4-20250514`、`claude-opus-4-1-20250805`、`claude-opus-4-6`、`claude-sonnet-4-20250514`、`claude-sonnet-4-6`
- 模型广场当前仍有旧品牌模型 `claude-opus-4.6`、`claude-sonnet-4.6`
- `GroupRatio` 当前仍是：
  - `claude_core = 1.5`
  - `gemini_core = 1`
- 使用日志前端会在映射场景下直接显示：
  - `请求并计费模型`
  - `实际模型`

## Design

### 1. 模型同步策略

这次不做“品牌别名包装”，直接以 `CLI API` 的真实模型名同步。这是最小改动方案，避免再引入一层新的品牌映射，影响当前套餐分组和计费路由。

处理方式：

- 删除旧的公开模型元数据：
  - `claude-opus-4.6`
  - `claude-sonnet-4.6`
- 保留并新增真实模型名：
  - GPT 系列统一进入 `gpt_core`
  - Claude 系列统一进入 `claude_core`
- `gpt-5.2 / gpt-5.2-codex` 现有 `gpt52_unlimited` 内部路由不动

这样做的结果是：

- 模型广场展示的是稳定的真实模型名
- 用户调用也是同名模型，不再依赖品牌别名
- 现有套餐“基础只能 GPT / Coding Plan 全开 / 5.2 系列无限”这套权限结构不需要重排

### 2. 定价来源

价格优先复用项目原生定价能力，而不是手写一套新表。

来源顺序：

1. 使用项目内置 `ratio_setting` 默认模型倍率
2. 将确认存在的新模型写入数据库当前使用的 `ModelRatio / CompletionRatio`
3. 对没有显式补全倍率的模型，沿用项目默认推导或与输入倍率一致的策略

这次只做“写入一版可用价格”，后续如果你在后台手工微调，不会与本次结构冲突。

### 3. 渠道路由与分组

现有自动路由结构保持不变：

- `AutoGroups = ["gpt52_unlimited", "gpt_core", "gemini_core", "claude_core"]`

新增模型只补到现有分组里，不新增新的套餐分组：

- GPT 新模型加入 `gpt_core`
- Claude 新模型加入 `claude_core`

删除旧品牌模型时，要一并清理它们对应的：

- `models` 元数据
- 渠道 `models`
- `model_mapping`
- 旧能力记录

但不能影响已有可用模型，尤其是：

- `gpt-5.4-thinking`
- `gpt-5.4-mini`
- `gpt-5.2`
- `gpt-5.2-codex`
- `gpt-5.3-codex`
- 现有 Gemini 系列

### 4. 分组倍率

按你的要求改成：

- `gpt_core = 1`
- `gemini_core = 1.5`
- `claude_core = 4`

保留：

- `gpt52_unlimited = 1`
- `plan_codex_advanced -> gpt52_unlimited = 0`
- `plan_coding_all -> gpt52_unlimited = 0`

这意味着：

- Claude 系列在调用时会按 4 倍消费
- Gemini 系列按 1.5 倍消费
- GPT 系列维持 1 倍
- `gpt-5.2 / gpt-5.2-codex` 在进阶和 Coding Plan 下仍然是 0 倍率

### 5. 日志隐私

这次不删后端审计数据，只隐藏用户可见展示。

处理方式：

- 前端使用日志列表不再展示“实际模型”映射 Popover
- 详情展开中不再追加：
  - `请求并计费模型`
  - `实际模型`
- 后端 `other.upstream_model_name` 暂时保留，方便管理员后续排障

这样做的好处：

- 用户端不会再看到真实映射后的上游模型
- 审计和服务端排障信息仍在，不会破坏日志结构
- 改动面最小，风险低于直接改日志写入链路

## Risks

### 1. 旧模型名兼容性

删除 `claude-opus-4.6` 和 `claude-sonnet-4.6` 后，任何仍然调用这两个旧名字的客户端都会失败。需要接受这两个旧名称正式下线。

### 2. 新模型价格完整性

项目原生倍率表覆盖了这批新模型的大部分名称，但个别模型若无显式补全倍率，可能需要沿用输入倍率或后续手动修正。这属于可控风险，因为你后面会继续人工检查价格。

### 3. 能力记录一致性

新增/删除模型时，要同步清理或补齐：

- `channels.models`
- `channels.model_mapping`
- `abilities`
- `models`

否则会出现：

- 模型广场能看到但不能调
- 或能调但模型广场无元数据

## Testing

需要覆盖 4 类验证：

1. 模型广场：
   - 不再出现 `claude-opus-4.6`、`claude-sonnet-4.6`
   - 出现新增的 GPT/Claude 新模型
2. 实际调用：
   - 新增 GPT 模型能成功调用
   - 新增 Claude 模型能成功调用
   - 原有套餐分组仍按预期限制模型可用范围
3. 计费倍率：
   - `claude_core = 4`
   - `gemini_core = 1.5`
   - GPT 保持 1
4. 使用日志：
   - 不再显示“实际模型”映射
   - 日志中的公开模型名保持用户请求名，而不是上游映射名
