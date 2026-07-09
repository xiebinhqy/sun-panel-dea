# 在线更新功能实施完成清单

## ✅ 1. 更新所有链接引用 (hslr-s → xiebinhqy/sun-panel-dea)
- [x] src/components/apps/About/index.vue - 作者、GitHub、Issues等所有链接
- [x] src/views/login/index.vue - 登录页脚链接
- [x] src/store/modules/panel/helper.ts - 默认页脚HTML
- [x] service/initialize/A_ENTER.go - 项目地址
- [x] src/components/apps/ImportExport/index.vue - 书签转换工具链接

## ✅ 2. 后端：添加更新检查与执行接口
- [x] 新增: service/api/api_v1/system/update.go - GitHub Release检查+下载更新API
- [x] 新增: service/router/system/update.go - 更新路由定义
- [x] 修改: service/api/api_v1/system/A_ENTER.go - 注册UpdateApi
- [x] 修改: service/router/system/A_ENTER.go - 注册更新路由

## ✅ 3. 前端：添加更新UI (关于页)
- [x] 新增: src/api/system/update.ts - 更新API调用
- [x] 修改: src/components/apps/About/index.vue - 添加在线更新按钮、弹窗、进度条
- [x] 修改: src/locales/zh-CN.json - 添加更新相关中文文案
- [x] 修改: src/locales/en-US.json - 添加更新相关英文文案
- [x] 修改: package.json - 更新name和version