pd - 静态博客生成工具
===

安装
---
编译需要配置好golang开发环境
	
	go get -u github.com/zhoukk/pd
	go install github.com/zhoukk/pd

或直接下载编译好的运行文件[pd.exe](https://github.com/zhoukk/pd/releases/)

命令
---

### pd new [sitename] 创建站点

在当前目录下生成名为sitename的目录做为博客主目录。在主目录下生成.pd目录和about.md文件。其中.pd为配置目录。about.md文件为生成关于页面的内容。

### pd post [name] [title] 创建新文章

在.pd/posts 目录下生成[name].md文件，并在文件中生成文章的元数据。

	{
		"date": "2006-01-02 15:04:05",
		"description": "pd生成的第一篇文章",
		"permalink": "/2006/01/first.html",
		"category": "默认",
		"title": "第一篇文章"
	}

	。。。以下为文章正文内容

编辑[name].md文件，修改以上的元数据。在下面以markdown格式编写文章内容。

### pd compile 编译站点

遍历.pd/posts 目录下所有markdown格式的文章，依次生成对应html文件，路径格式为/ year/month/name.html。拷贝主题所需资源文件到主目录下static目录中。为站点生成rss.xml，atom.xml，sitemap.xml 文件。


### pd http [port] 预览站点

启动测试web服务器，默认port为:80。启动浏览器，输入http://127.0.0.1预览。

### pd update 更新站点

更新站点目录下.pd目录中的主题目录theme。以支持最新功能。


配置
---

在站点主目录下的.pd目录内有config.json配置文件。

	{
		"domain": "http://domain.com",
		"author": "author",
		"keywords": "keywords",
		"description": "description",
		"title": "title",
		"ajax": "http://ajax.domain.com",
		"blog_per_page": 10,
		"pagination_show_num": 7,
		"summary_line": 20,
		"theme": "sample"
	}

- domain
站点的域名，用于生成rss.xml，atom.xml，sitemap.xml等
- author
站点作者
- keywords
站点关键字
- description
站点描述
- title
站点主页的title
- blog_per_page
每页文章数，用于分页
- pagination_show_num
分页栏显示页数，用于分页
- theme
站点使用的主题
- ajax
评论功能使用ajax请求的域名

评论
---
关于config.json中的ajax配置，此配置用于支持文章评论功能。评论内容支持markdown语法。配置的域名需提供以下两个接口：

- get ajax.domain.com/comment.list?id=xxx.html

	请求指定文章的评论列表，id为文章的uri部分。
	例如http://ajax.domain.com/comment.list?id=%2F2006%2F1%2Ffirst.html
	请求/2006/1/first.html文章的评论列表，返回

		[{
		"nickname":"user1",
		"url":"http://xxx.com",
		"content":"\u003cp\u003euser1@http://xxx.com 发表评论1\u003c/p\u003e\n",
		"time":"2015-11-17 15:51:53"
		},
		{
		"nickname":"user2",
		"url":"http://xxx.com",
		"content":"\u003cp\u003euser1@http://xxx.com 发表评论1\u003c/p\u003e\n",
		"time":"2015-11-17 15:52:00"
		},
		{
		"nickname":"user1",
		"url":"http://xxx.com",
		"content":"\u003cp\u003euser1@http://xxx.com 发表评论2\u003c/p\u003e\n",
		"time":"2015-11-17 15:52:07"
		}]

	
- post ajax.domain.com/comment.new

	对指定文章发表评论，数据为
	
		id:文章uri
		nickname:昵称
		url: nickname链接
		content: 内容
	
例如：

	id=%2F2006%2F1%2Ffirst.html&nickname=user1&url=http%3A%2F%2Fxxx.com&content=评论内容
