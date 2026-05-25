---
name: LLM Gateway Portal
description: A calm, distinctive workspace for managing API access and usage.
colors:
  portal-canvas: "#F4F9FC"
  portal-surface: "#FCFEFF"
  portal-ink: "#223842"
  portal-muted: "#60757D"
  portal-border: "#D9E6EC"
  portal-primary: "#237BB2"
  portal-primary-hover: "#1A6392"
  portal-highlight: "#E7F2F8"
  portal-accent: "#6FAECD"
  portal-accent-deep: "#266889"
  portal-accent-soft: "#EAF5FB"
  portal-danger: "#B94B4C"
typography:
  headline:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "32px"
    fontWeight: 650
    lineHeight: 1.2
  title:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "20px"
    fontWeight: 600
    lineHeight: 1.35
  body:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.65
  label:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "12px"
    fontWeight: 600
    lineHeight: 1.4
rounded:
  sm: "8px"
  md: "12px"
  lg: "18px"
spacing:
  xs: "8px"
  sm: "12px"
  md: "16px"
  lg: "24px"
  xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.portal-primary}"
    textColor: "{colors.portal-surface}"
    rounded: "{rounded.sm}"
    padding: "0 18px"
    height: "42px"
  portal-panel:
    backgroundColor: "{colors.portal-surface}"
    textColor: "{colors.portal-ink}"
    rounded: "{rounded.lg}"
    padding: "24px"
---

# Design System: LLM Gateway Portal

## 1. Overview

**Creative North Star: "The Quiet Ledger"**

门户是一张整理清楚的访问工作台：清浅蓝白画布和稳定内容表面承载调用信息，天蓝色引导关键操作，提示区域以近乎水洗的浅蓝承接安全信息。它服务于凭据与调用信息，因此应精确、轻松扫描，同时保留足够呼吸空间。

系统拒绝“传统管理后台模板式的深色侧栏加白色表格堆叠”，也不采用终端式深色界面、荧光强调色或装饰性玻璃效果。辨识度来自克制的构图、自然色调和敏感操作中的细致反馈。

**Key Characteristics:** 清浅蓝白画布；统一天蓝信号；任务优先的内容层级；适配键盘与移动触控的明确交互。

## 2. Colors

色板以带水汽感的浅蓝白画布承载深青灰文字。可读的天蓝用于路径与行动，提示区仅使用极浅同色系底色；颜色更统一，减少不同色相之间的冲突。

### Primary
- **Sky Action Blue** (`#237BB2`): 主按钮、选中导航和焦点状态，使用可访问范围内更明亮的一档天蓝。

### Secondary
- **Washed Sky Accent** (`#6FAECD`): 非正文的小面积装饰性蓝色。
- **Deep Sky Label** (`#266889`): 安全提示图标与上眉标签等需要可读的小面积强调。

### Neutral
- **Air Canvas** (`#F4F9FC`): 整体页面背景。
- **Clear Surface** (`#FCFEFF`): 面板、导航和表单所在表面。
- **Deep Water Ink** (`#223842`): 标题与主要正文。
- **Cloud Note** (`#60757D`): 辅助信息。
- **Blue Mist Rule** (`#D9E6EC`): 分割和面板边界。

**The One Hue Rule.** 天蓝用于操作与选中状态，浅蓝只承载提示背景；不再引入第二种品牌色相。

## 3. Typography

**Display Font:** Inter (with system sans-serif fallback)  
**Body Font:** Inter (with system sans-serif fallback)

**Character:** 一套稳定的人机界面字体承担所有内容，利用字重、间距和数字对齐提供清晰的信息层次。

### Hierarchy
- **Headline** (650, 32px, 1.2): 页面唯一主标题。
- **Title** (600, 20px, 1.35): 面板标题与关键区域名称。
- **Body** (400, 14px, 1.65): 说明文字，长段内容控制在 70 字符左右。
- **Label** (600, 12px, 1.4): 上眉标题、状态和次要元数据。

## 4. Elevation

系统以色调分层和极轻的环境阴影建立深度。常态容器主要依赖边界和表面对比，悬浮状态只作有限提升，不制造漂浮卡片墙。

### Shadow Vocabulary
- **Panel Ambient** (`0 1px 2px rgba(34, 56, 64, 0.03), 0 12px 30px rgba(32, 83, 118, 0.04)`): 门户主要面板。
- **Header Float** (`0 1px 0 rgba(34, 56, 64, 0.05)`): 顶部栏与正文分界。

## 5. Components

### Buttons
- **Shape:** 轻圆角矩形 (`8px`)，最小触控高度 `42px`。
- **Primary:** Sky Action Blue 表面、Clear Surface 文字，仅用于页面主要行为。
- **Hover / Focus:** 加深主色并显示可见焦点环，状态过渡约 `180ms`。
- **Secondary / Ghost:** 使用透明或白色表面与柔和边界，不与主动作争抢层级。

### Cards / Containers
- **Corner Style:** 宽松圆角 (`18px`)。
- **Background:** Clear Surface 位于 Air Canvas 上。
- **Shadow Strategy:** 仅使用 Panel Ambient。
- **Border:** Sea Glass Rule `1px` 边界。
- **Internal Padding:** 桌面 `24px`，窄屏 `16px`。

### Inputs / Fields
- **Style:** 浅色实体输入面，`8px` 圆角和柔和边界。
- **Focus:** Sky Action Blue 焦点环。
- **Error / Disabled:** 文字说明配合状态色，不能只依赖颜色。

### Navigation
- 顶部轻量工作台导航替代传统深色侧栏；当前区域使用浅蓝选中面和天蓝文字，提示区域也仅使用更轻的蓝色底。小屏下布局换行并保持操作目标尺寸。

## 6. Do's and Don'ts

### Do:
- **Do** 在门户页面以 `#F4F9FC` 画布、`#FCFEFF` 面板和 `#237BB2` 主要操作形成稳定层级，以 `#EAF5FB` 承载提示背景。
- **Do** 优先说明 Key 安全保存、额度与日志等用户实际任务。
- **Do** 为键盘焦点、加载、空数据和窄屏表格提供明确处理。

### Don't:
- **Don't** 回到“传统管理后台模板式的深色侧栏加白色表格堆叠”。
- **Don't** 使用偏开发者终端的黑底、荧光色和代码编辑器气质作为门户默认外观。
- **Don't** 使用装饰性玻璃效果、渐变文字、彩色侧边条或无意义动效制造独特感。
