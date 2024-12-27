# 调度框架日志模块

#### 日志模块的使用

1.   导入pkg下的日志模块：`pkg/component-base/logs`
2.   对日志模块进行初始化，具体初始化方式： `logs.Init(moduleName)` ， 同一个模块只初始化一次即可
3.   通过`logs.Trace() \ logs.Debug() \ logs.Info() \ logs.Warn() \ logs.Error() \ logs.Fatal() \ logs.Panic() `调用不同级别的日志，它们可以接受的参数类型为：`任意的结构体类型、任意的字符串、以及map[string]interface{}类型`，可以接受任意多的参数
4.   通过`logs.Tracef() \ logs.Debugf() \ logs.Infof() \ logs.Warnf() \ logs.Errorf() \ logs.Fatalf() \ logs.Panicf() `可以接受类似于Printf接受的格式化字符串

#### 日志的存储位置

日志分模块存储在`/tmp/logs/`目录下

-   ​	比如在模块A初始化日志`logs.Init(“A”)`， 则它的日志都会存储在`/tmp/logs/A/`目录下
