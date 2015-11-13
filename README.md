pd - golang实现的静态博客生成工具
========

安装
===

	go get -u github.com/zhoukk/pd
	go install github.com/zhoukk/pd

配置
===

config.json

1. domain 		网站域名
2. author 		网站作者
3. theme  		网站布局颜色主题
4. title  		主页title
5. description 	主页描述
6. port 		测试端口
7. root 		网站目录

命令
===

1. pd new
2. pd compile
3. pd http
4. pd help

新建站点
===

1. 新建站点目录
2. 配置config.json中的root目录 及其他
3. 照片目录/photos, 照片缩略图目录/photos/thumb
4. 视频目录/videos
5. 文章目录/posts

新建文章
===

1. pd new 文章名字  会在root/post/目录下生成md文件,头部生成meta元数据
2. 编辑root/post/文章名字.md 文件

	1. data 		文章生成日期
	2. description	文章描述
	3. permalink	文章链接
	4. category		文章类目
	5. title 		文章标题

3. 以markdown格式在后面编写文章内容

编译站点
===

	pd compile

1. 读取配置文件config.json
2. 加载主题文件
3. 读取照片目录root/photos所有照片
4. 读取视频目录root/videos所有视频
5. 读取文章目录root/posts所有文章
6. 生成每篇文章的html文件
7. 生成网站其他html文件
8. 拷贝主题目录下资源文件夹/js /css /img /fonts到/root/static目录
9. 生成rss.xml atom.xml sitemap.xml

本地预览
===

	pd http

启动本地http服务，根目录为config.json -> root目录
http://127.0.0.1 查看
