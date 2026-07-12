# 功能实现计划

## 功能1：将右下角按钮移至右上角横向排列

**文件：** `src/views/home/index.vue`

- `.fixed-element` CSS: `bottom: 50px` → `top: 10px`, `right: 10px`
- `NButtonGroup` 从 `vertical` 改为横向
- 按钮顺序：网络切换按钮排前，系统应用按钮排后

## 功能2：系统应用弹窗可拖拽 + 悬停特效

**文件：** `src/components/common/DraggablePanel/index.vue`（新建）
**文件：** `src/views/home/components/AppStarter/index.vue`（修改）

- 创建 DraggablePanel 组件：固定定位、双击/拖拽、可自由移动
- AppStarter 使用 DraggablePanel 替代 RoundCardModal
- 左侧应用列表菜单项：hover 时加粗 + box-shadow 特效