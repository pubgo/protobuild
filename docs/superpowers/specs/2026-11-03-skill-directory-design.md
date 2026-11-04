# 设计：protobuild Agent Skill 以仓库目录分发，删除 skills 子命令

- 日期：2026-11-03
- 分支：`chore/agents-skills`
- 状态：已批准（2026-11-03），实施见 docs/superpowers/plans/2026-11-03-skill-directory.md

## 意图与背景（Why）

protobuild 用户希望他们的编码智能体"知道怎么正确使用 protobuild"——不只是执行命令，还包括做判断：没有 `protobuf.yaml` 时先 init、改 proto 后知道跑 breaking、出错时用 doctor 诊断、知道 `.proto/` 不能手改这类项目坑。

现有对外通道是 `protobuild skills` 子命令（把 `cmd.go` 内联字符串模板写到用户项目 `.agents/skills/protobuild/`）。用户判断：**该通道多此一举且不完整**（命令列表已与实际漂移：模板缺 `init`/`skills`；`docs/AGENT_SKILL.md` 的 enum 更旧）。决定改为：**技能以仓库内真实目录为唯一真源**，用户经 `npx skills add pubgo/protobuild`（支持 `-g` 全局）安装；`protobuild skills` 子命令删除。

## 目标

1. 新增 `skills/protobuild/` 目录（入库、可评审、可 diff）：
   - `SKILL.md`：frontmatter 遵循 Agent Skills 规范（name、description 等），正文走渐进披露，保持精短（约 100 行内），详表链接到 references。
   - `references/commands.md`：全命令 + 关键 flag 速查（以 `--help` 实际输出为准：init / vendor / update / verify / gen / buf / breaking / lint / format / deps / install / clean / doctor / web / version / upgrade）。
   - `references/config.md`：面向智能体的操作要点（裁剪、自包含）：`protobuf.yaml` 关键字段与 deps 多源（git/gomod/http/s3/gcs/local）要点、`protobuf.plugin.yaml` 目录级覆盖；人类向完整教程保留在 `docs/EXAMPLES.md`，不双抄细节。
   - `references/troubleshooting.md`：doctor / verify 用法 + 常见故障与修法。
   - 内容覆盖全流程：安装前提（protobuild、protoc、插件）、init 与配置引导、vendor→gen 常规流程、lint/format/breaking 检查、诊断。
   - 安全约束：写盘命令（gen、format -w、install、buf、clean）需告知用户；须在有 `protobuf.yaml` 的项目根执行；先 vendor 再 gen；`.proto/` 禁止手改（重建用 `vendor -f`）；buf 只认 `buf.yaml`。
   - `agents/openai.yaml`：保留原 openai 元数据内容于技能目录内（供 OpenAI 兼容平台），但去掉 icon_small / icon_large 字段（仓库无图片资产，原样保留即坏引用）。〔我方拍板〕
   - frontmatter `description` 触发范围：不限死"仅当 protobuf.yaml 存在"（否则用户要求的 init/创建/诊断场景永不触发），改为"用 protobuild 管理 protobuf 工程时：存在 protobuf.yaml，或用户要求搭建/初始化/诊断 protobuild"。〔builder-check 修订〕
2. 删除 `cmd/protobuild/cmd.go` 中：`defaultSkillContent`、`defaultOpenAIContent`、`newSkillsCommand` 及其在 `Children` 的注册、`normalizeTemplateIndent`、`writeFileIfNeeded`（仅此命令使用，已 grep 确认无其他引用、无测试）。
3. 文档同步（一次性消灭已知漂移）：
   - `README.md`：删命令表中 `skills` 行（第 109 行附近）。
   - `AGENTS.md`："新增子命令三件套"改为两件套 + 同步更新 `skills/protobuild/` 文档；并把 `AGENTS.md` 本身提交入库（当前 untracked）。
   - `docs/AGENT_SKILL.md`：接入方式改写为 `npx skills add pubgo/protobuild`（含全局安装）与直接拷贝目录两条路径；命令 enum 不再手抄，改为指向 `skills/protobuild/SKILL.md`；补一段迁移提示：旧 `protobuild skills` 已移除，此前生成的 `.agents/skills/protobuild/` 请删除或覆盖，避免漂移旧内容遮蔽新装技能；保留"推荐调用策略/最小操作清单"精神（写盘提示、CI 用 `format --exit-code` + `lint`）。
   - `docs/INDEX.md`：如有需要补链接（构建时核对）。
4. 防漂移守卫测试（`cmd/protobuild` 包内）：遍历 `Main().Children`，断言每个子命令的 `Use` 名出现在 `skills/protobuild/references/commands.md`。纯文档一致性断言，不改任何命令行为。

## 非目标

- 不重构 `docs/`（EXAMPLES/MULTI_SOURCE_DEPS 等保持原位；重叠内容按受众切，不做 docs 大搬家——路线 B 未选）。
- 不为已删除的 `skills` 子命令做 tombstone 或新代码提示。
- 不改动其他子命令的任何行为。
- 不改 web UI。
- 不发布到第三方 skills 市场；仓库本身即通道（`npx skills add` 直接拉仓库）。
- 不为 protobuild 自身的开发流程另写技能（AGENTS.md 已承担）。

## 反目标（做成了也算失败的情况）

- 又出现"文档一份、代码一份"双真源——正是此次废弃的病根。
- SKILL.md 变成无结构的中英混杂堆料，智能体无法扫读。
- 内容与仓库实际命令/flag 不符（新漂移）。

## 约束

- 删除子命令是对外的破坏性变更：已获用户明确决策，接受。
- `skills/protobuild/` 需能通过 `skills-ref validate`（本地无网络时以规范自查代替，CI 不强求）。〔我方拍板〕
- 命令与 flag 事实必须取自仓库源码 `--help` 与 docs，不得凭记忆编造。
- 全仓测试保持绿（`go test ./...`）；`task all` 本地全绿再提 PR。

## 方案比较（用户已选 C）

- **A：目录入库 + go:embed 让 `protobuild skills` 写出同一文件**。双通道同源，但要动代码、embed 与测试。用户否：通道本身多此一举。
- **B：目录入库、CLI 不动**。最快且纯文档，但双源并存必然再漂移。否。
- **C：目录入库、废弃子命令（选定）**。单一真源，可评审，`npx skills add pubgo/protobuild`（含 `-g`）拷贝即用。

## 留给构建者的

- SKILL.md 正文分节（建议：When to use / Prerequisites / Workflows：init、vendor+gen、检查（lint/format/breaking）、诊断（doctor/verify）/ Safety / Command index）。
- references 三文件的详略与拆分粒度；若单文件超约 200 行可再拆。
- 语言：正文中文、frontmatter description 英文为主（用户群以中文为主，agent 检索以英文更稳）。〔我方拍板〕

## Builder-check 处置记录

构建者审读提出 8 问，两处需要你拍板（见下"待定"），其余处置如下：

- 触发限制 vs init 矛盾：已修订（见目标 1），否则 init/诊断目标永不触发。
- openai.yaml 图片引用悬空：去掉 icon 字段（见目标 1）。
- 版本号：本 PR 不动 `.version/VERSION`（0.x 破坏性变更的升位按发版流程走：改 VERSION → 合并 → 打 tag）。〔我方拍板〕
- 删除形态：纯删除，不做 tombstone；迁移提示只写进 `docs/AGENT_SKILL.md`（见目标 3），不为已删命令维护第二条帮助文本。〔我方拍板〕
- `skills/` 在本仓库自发现的副作用：接受——AGENTS.md 仍是开发语境权威，技能内容（面向 protobuild 使用者）与之不冲突。〔我方拍板〕
- 本设计文档与后续计划文档随本 PR 入库。〔我方拍板〕

## 已定（用户选择推荐项）

1. 真源切分走 **C**：按受众切——智能体操作所需（命令速查、安全约束、诊断流程、配置要点）收进 `skills/protobuild/references/` 做单一真源；`docs/` 人类文档保留教程/设计叙述定位，重叠面压到最小。
2. **防漂移守卫纳入本 PR**：`cmd/protobuild` 包内加测试，遍历 `Main().Children` 断言每个子命令名出现在 `references/commands.md`，漏同步文档时 CI 直接红。

## 验收

1. `go test ./...`、`golangci-lint run`、`task all` 全绿。
2. `protobuild` 的 `--help`/`Children` 中不再出现 `skills`；全仓 grep 无 `defaultSkillContent` 残留。
3. `skills/protobuild/SKILL.md` frontmatter 完整、正文含上述全流程要点与 Safety；references 与真实 `--help` 一致（抽查 `init --template`、`doctor --fix`、`format --exit-code`、`vendor -f`）。
4. README/AGENTS.md/docs/AGENT_SKILL.md 三处文档无 `skills` 子命令残留、无旧 enum 漂移；AGENTS.md 已跟踪。
5. `npx skills add pubgo/protobuild` 的说明写入 AGENT_SKILL.md（安装指令本身由生态工具负责，不新增代码）。
6. 防漂移守卫真实生效：临时从 `references/commands.md` 删一个命令名，测试应失败；恢复后全绿（验证非摆设）。
