# CrashService

UE相关平台Crash服务处理，包含客户端（windows,android,ios）和服务端（windows（minidump）,linux）

已完成支持功能：<br>
1) Windows Crash上报收集和解析
2) UE原生的CrashReportClient的收集器，作为CrashReportServer（不用和CrashHandle配合使用）
2) Error 错误日志上报和收集
3) 符号表上传和接收
4) Linux Crash 收集，Linux 上报方式和 Windows 一样
5) Android Crash 收集和解析

未完成功能：<br>
1) IOS Crash收集和解析（上报依赖于CrashHandle插件）
2) Mac Crash收集上报和解析
3) Crash 通知（飞书等）
4) 手工解析，补漏

该功能必须配合UE4插件 https://github.com/qqwx1986/CrashHandle 一起使用

## 客户端
### windows 上报 Minidump

执行 build_client_windows.bat 生成到 build/CrashReportClient.exe ,或者直接下载已经编译好的二进制文件<br>

打包UE4项目后，拷贝 build/CrashReportClient.exe 到 ${ClientPack}/Engine/Binaries/Win64/

### 编辑器原生上报
修改 Engine/Programs/CrashReportClient/Config/DefaultEngine.ini

DataRouterUrl="http://127.0.0.1:13333/receiverUECrash"

### linux 上报崩溃日志

目前linux一般指的是UE的dedicated server，打包发布时建议符号表一起发布，就是 *.sym 和 *.debug 一起发布，<br>
这样DS崩溃时直接会生成崩溃堆栈到日志文件，这边只需上报日志文件即可

执行 build_client_linux.bat 生成到 build/CrashReportClient ,或者直接下载已经编译好的二进制文件<br>
打包UE4项目后，拷贝 build/CrashReportClient 到 ${ClientPack}/Engine/Binaries/Linux/

### android 上报
https://github.com/qqwx1986/CrashHandle 中上报处理

### 上传符号表
执行 build_client_windows.bat 生成到 build/CrashReportClient.exe ，或者执行 build_client_linux.bat，build_client_linux.sh，上传符号表的客户端可以跑在win&linux下

-url 上传httpUrl ${httpUrl}

-symbolPath 符号表路径 ${symbolPath}

-executePath 二进制文件路径 ${executePath}，改参数 platform 为 android 时不需要

-platform 符号表的平台类型（windows/linux/android）${platform},windows下就是pdb和exe

-version 符号表的版本号，如果不填，android默认是 latest，linux和windows默认是二进制执行文件的md5码

```shell
CrashReportClient.exe uploadSymbol -url=${httpUrl} -symbolPath= ${symbolPath} -executePath=${executePath} -platform=${platform}
    
./CrashReportClient uploadSymbol -url=${httpUrl} -symbolPath= ${symbolPath} -executePath=${executePath} -platform=${platform}
```
### 上传错误日志
配合  https://github.com/qqwx1986/CrashHandle 插件，调用以下接口上传错误信息
```shell
void UCrashHandleBlueprintLibrary::ReportError(const FString& Error);
```

## 服务器

### 搭建收集dump处理服务器（windows,linux,android），备注：处理windows的minidump只能在windows系统上运行
 
执行 build_server_windows.bat 生成到 build/CrashReceiverServer.exe ,或者直接下载已经编译好的二进制文件<br>

拷贝 build/CrashReceiverServer.exe 到 ${ServerDir}

下载 https://raw.githubusercontent.com/qqwx1986/DumpAnlyise/main/Release/DumpAnlyise.exe 到 ${ServerDir}/tool/，这一步骤可以略过，会自动下载

-httpPort http端口 

-basePath 工作目录 ${basePath}，会在该目录下创建 crash&symbol&error&tool 三个目录，分别放置收集的crash文件，符号表，和收集的错误日志

-dumpAnlyiseUrl DumpAnlyise.exe下载的url ${dumpAnlyiseUrl}，默认已经填了github官方下载地址，所以这个选项没有特殊情况不需要，自动下载到 ${basePath}/tool/ 目录下

-chanNum 并发解析dump的channel数（默认1），单个任务消耗400MB，根据系统内存来分配大小

-daemon 是否后台进程

-addr2linePath addr2line的目录（为解析android和linux符号表的工具），如果运行在linux下，该参数不用填，如果运行在windows下，下载addr2line.exe到 ${basePath}/tool/ 下

-androidSymbolName 为解析 android 的符号表名称，UE4打完包后 ${UE4_Project}/Binaries/Android/ 目录下 *.so 的具体名称，上传的符号表也是这个

```shell
CrashReceiverServer.exe server -httpPort=13333 -basePath=${basePath} -dumpAnlyiseUrl=${dumpAnlyiseUrl}  -chanNum=2 -daemon=false
```
### 符号表接收服务
执行 build_server_windows.bat 生成到 build/CrashReceiverServer.exe，也可以执行 build_server_linux.bat，build_server_linux.sh ,该服务可以跑在windows&linux下<br>
参数和堆栈解析服务一样，同一个进程服务可以同时接收符号表和解析堆栈<br>
该服务涉及符号表的接收的压缩也解压缩，接收符号表时需要消耗5GB内存

### 接收错误日志服务
执行 build_server_windows.bat 生成到 build/CrashReceiverServer.exe，也可以执行 build_server_linux.bat，build_server_linux.sh ,该服务可以跑在windows&linux下<br>
参数和堆栈解析服务一样，同一个进程服务可以同时接收符号表和解析堆栈和接收错误日志


### 浏览
http://${httpUrl}/error 浏览错误信息

http://${httpUrl}/crash 浏览崩溃信息

http://${httpUrl}/UECrash 浏览UE崩溃信息

http://${httpUrl}/symbol 浏览符号表

以本地启动示例 http://localhost:13333/error

## 二进制下载
https://github.com/qqwx1986/CrashService/tree/main/build
