# UfopFrame

## 简介
ufopframe 是为七牛 ufop 开发的一个框架，通过修改其中一些代码，可以轻松的开发部署自己的 ufop 服务

## 使用步骤
* 下载框架代码，查看框架的目录结构
	
	```
	.
├── CHANGELOG.md
├── LICENSE
├── README.md
└── ffmpeg
    ├── deploy
    │   ├── Dockerfile
    │   ├── dora.yaml
    │   └── ffmpeg
    │       ├── ffmpeg.conf
    │       ├── qufop
    │       └── qufop.conf
    └── src
        ├── build.sh
        ├── cross_build.sh
        ├── ffmpeg.conf
        ├── qufop.conf
        ├── qufop.go
        └── ufop
            ├── base.go
            ├── config.go
            ├── ffmpeg
            │   └── ffmpeg.go
            ├── server.go
            └── utils
                ├── charset.go
                ├── param.go
                ├── reqid.go
                └── tools.go
	```
	* ``src`` 是代码目录
		* ``build.sh``: 是当前环境的编译脚本，可以直接使用下面的命令编译当前项目
		
			```
			>>./build.sh
			```
			 
		* ``cross_build.sh``: 是 ``linux`` 环境的编译脚本，使用方式同上
		* ``ffmpeg.conf``: 是当前项目需要的配置，比如鉴权需要的 ak/sk 等
		* ``qufop.conf``: 是框架所需要的配置，比如一下联网配置，ufop 前缀等
		* ``qufop.go``: 是入口函数，配置了一些常规的日志输出方式，读取配置文件等
		* ``base.go``: 定义一些常量和框架本身的配置文件，主要是定义了 ``UfopJobHandler`` 接口
		* ``config.go``: 配置文件加载类 
		* ``server.go``: 服务类，主要包括创建/启动server，格式化返回内容等
		* ``utils``: 主要是一些基础工具类，包括解析参数，生成随机 reqid 等
		* ``ffmpeg``: 主要用来实现 ufop 等功能，里面的类需要继承 ``base.go`` 中的 ``UfopJobHandler`` 接口
	* ``deploy`` 是 Ufop 的部署目录
		* ``Dockerfile``: 用来生成 docker 镜像
		* ``dora.yaml``: 七牛 ufop 的配置文件，可以通过工具来动态获取，获取后修改相应的参数即可
		* ``ffmpeg``: ufop 的主要处理目录，需要用到的工具都可以放到这个目录下

* 修改相应的框架代码，生成自己的项目框架
	* 修改 `ffmpeg.conf`，生成自己项目需要的配置文件，如果修改了配置文件的名字，同时需要修改 `qufop.go` 中代码
		
		```
		//register job handlers
	if err := ufopServ.RegisterJobHandler("ffmpeg.conf", &ffmpeg.FFmpeg{}); err != nil {
		log.Error(err)
	}
		```
	* 修改 ``ffmpeg``目录下的 ``ffmpeg.go``代码
	
		```
		1. FFmpeg 主类，继承 base.go 中的 UfopJobHandler 接口
		2. Config 类，用来解析项目配置文件
		3. InitConfig 方法，解析项目配置文件
		4. parse 私有方法，主要用来解析 cmd 中的参数
		5. Do 方法，项目的主要处理方法，可以在这里实现 ufop 的功能
		```
* 编译代码，生成部署目录
	* 将编译好的可执行文件，配置文件以及其它依赖的文件放到 `deploy` 的 `ffmpeg` 下
	* 修改 `Dockerfile` 文件，生成 `docker` 镜像
	* 下载七牛官方的 `qufop` 工具，[下载地址](https://developer.qiniu.com/dora/tools/1222/qdoractl)
	* 参考文档，将 `docker` 镜像推到七牛 `ufop` 平台
	* 开始测试