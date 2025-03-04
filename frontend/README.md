# AI Proxy Go 前端目录

## 目录结构
- static/: 静态资源文件
  - css/: 样式文件
  - js/: JavaScript文件
  - images/: 图片资源
- templates/: HTML模板文件
  - install/: 安装相关页面
  - login/: 登录相关页面
  - home/: 主页相关页面

## 开发规范
1. 静态文件命名规则：
   - CSS文件：模块名.css
   - JS文件：模块名.js
   - 图片文件：用途_尺寸.png/jpg

2. 模板文件规范：
   - 每个模块一个目录
   - 主页面命名为index.html
   - 子页面使用功能名称命名

3. 资源引用规范：
   - CSS文件通过/static/css/路径引用
   - JS文件通过/static/js/路径引用
   - 图片通过/static/images/路径引用 