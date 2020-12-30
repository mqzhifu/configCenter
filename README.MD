个人-目的
----------
>熟悉一下GO，当练个手

>所有项目都需求配置中心这个东西，所以自研个简单的

配置中心-解决问题
-------
>1、传统开发，配置文件与项目混在一起，或者区分个文件夹，不易于管理与CICD

>2、不支持热变更

>3、不支持watch

>4、不容易做分布式

>5、在项目中修改配置文件，也算是修改了项目代码，容易影响到正常的业务代码

理想中的设计，模块拆分
-----------------
>1、SAPI，支持使用者可根据自定义协议查询/获取配置信息，如：TCP GRPC HTTP THRIFT等

>2、多类型的，解析文件，如：INI XML YML

>3、多类型的输出格式，如：JSON XML

>4、多层级的目录结构，如：N级文件夹

>5、配置文件层级遍历与内存持久化

>6、支持分布式与部署


目前实现的功能
--------
>1、支持HTTP

>2、层级仅支持：1级文件夹+1级文件

>3、文件类型解析，仅支持，.ini

>4、文件内容解析仅支持：.ini 中，模式(方括号)与KV

>5、输出格式，仅支持json


![image](https://www.processon.com/view/link/5f35257307912920b49aa2d1)

使用手册
--------
基础使用
```
var configerObj* Configer
#读取所有配置文件，持久化到内存的，上限值
#单个文件内容最大值
#所有配置文件个数，上限
#配置文件扩展名类型
configerObj = NewConfiger(128,1,1000,"ini")
#开启加载配置文件到内存中
configerObj.StartLoading("/data/www/golang/src/configCenter")
#搜索一个路径
configerObj.Search("/pj/main/secation_1/a")
```

开始HTTPD  SAPI
```
httpd = NewHttpd(3001,"127.0.0.1")
httpd.start()
```
