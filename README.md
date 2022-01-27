# CrashService

UE相关平台Crash服务处理，包含客户端（windows,android,ios）和服务端（windows（minidump）,linux）

已完成支持功能：<br>
1) Windows Crash上报收集和解析
2) Error 错误日志上报和收集
3) 符号表上传和接收

未完成功能：<br>
1) Linux Crash收集上报和解析（上报也可以使用CrashHandle插件）
2) Android Crash收集和解析（上报依赖于CrashHandle插件）
3) IOS Crash收集和解析（上报依赖于CrashHandle插件）
4) Mac Crash收集上报和解析
5) Crash 通知（飞书等）

该功能必须配合 https://github.com/qqwx1986/CrashHandle 一起使用
##客户端
## windows 处理 Minidump

执行 build_client_windows.bat 生成到 build/CrashReportClient.exe <br>

打包UE4项目后，拷贝 build/CrashReportClient.exe 到 ${ClientPack}/Engine/Binaries/Win64/

拷贝 config/CrashReportClient.json 到 ${ClientPack}/Engine/Binaries/Win64/ 并修改相关配置

Url 服务器httpUrl Ver 客户端二进制文件的版本（建议直接用二进制文件的md5作为版本号），如果填空，启动时会去
${ClientPack}/${ProjectName}/Binaries/Win64/下匹配相关二进制并实时计算md5码

###上传符号表
执行 build_client_windows.bat 生成到 build/CrashReportClient.exe ，或者执行 build_client_linux.bat，build_client_linux.sh，上传符号表的客户端可以跑在win&linux下

-url 上传httpUrl ${httpUrl}

-uploadSymbol 符号表路径 ${uploadSymbol}

-executePath 二进制文件路径 ${executePath}

-platform 符号表的平台类型（windows/linux/android）${platform},windows下就是pdb和exe
```azure
CrashReportClient.exe uploadSymbol -url=${httpUrl} -symbolPath= ${uploadSymbol} -executePath=${executePath} -platform=${platform}
    
./CrashReportClient.exe uploadSymbol -url=${httpUrl} -symbolPath= ${uploadSymbol} -executePath=${executePath} -platform=${platform}
```
### 上传错误日志
配合  https://github.com/qqwx1986/CrashHandle 插件，调用以下接口上传错误信息
```azure
void UCrashHandleBlueprintLibrary::ReportError(const FString& Error);
```

##服务器
###搭建 windows minidump服务器（只能跑windows系统）
 
执行 build_server_windows.bat 生成到 build/CrashReceiverServer.exe <br>

拷贝 build/CrashReceiverServer.exe 到 ${ServerDir}

下载 https://raw.githubusercontent.com/qqwx1986/DumpAnlyise/main/Release/DumpAnlyise.exe 到 ${ServerDir},

-httpPort http端口 

-basePath 工作目录 ${basePath}，会在该目录下创建 crash&symbol&error三个目录，分别放置收集的crash文件，符号表，和收集的错误日志

-dumpAnlyiseUrl DumpAnlyise.exe下载的url ${dumpAnlyiseUrl}，默认已经填了github官方下载地址，所以这个选项没有特殊情况不需要

-chanNum 并发解析dump的channel数（默认1），单个任务消耗400MB，根据系统内存来分配大小

-daemon 是否后台进程

```azure
CrashReceiverServer.exe server -httpPort=13333 -basePath=${basePath} -dumpAnlyiseUrl=${dumpAnlyiseUrl}  -chanNum=2 -daemon=false
```
### 符号表接收服务
执行 build_server_windows.bat 生成到 build/CrashReceiverServer.exe，也可以执行 build_server_linux.bat，build_server_linux.sh ,该服务可以跑在wind&linux下<br>
参数和堆栈解析服务一样，同一个进程服务可以同时接收符号表和解析堆栈<br>
该服务涉及符号表的接收的压缩也解压缩，接收符号表时需要消耗5GB内存

### 接收错误日志服务
执行 build_server_windows.bat 生成到 build/CrashReceiverServer.exe，也可以执行 build_server_linux.bat，build_server_linux.sh ,该服务可以跑在wind&linux下<br>
参数和堆栈解析服务一样，同一个进程服务可以同时接收符号表和解析堆栈和接收错误日志


### 浏览
http://${httpUrl}/error 浏览错误信息

http://${httpUrl}/crash 浏览崩溃信息

http://${httpUrl}/symbol 浏览符号表

以本地启动示例 http://localhost:13333/error

## 二进制下载
https://github.com/qqwx1986/CrashService/tree/main/build