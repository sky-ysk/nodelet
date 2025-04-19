# 序列化、反序列化、格式验证

```
导入模块 import "hit.edu/framework/pkg/component-base/analyzer"
导入接口定义相关包 import cores "hit.edu/framework/pkg/apis/cores" 
```

## 1. 序列化（serialize.go)

序列化提供了两个接口，一个是序列化为json文件，另一个是序列化为yaml文件，他们能接受任意结构体作为参数，调用的示例如下：

```go
// 输入参数： node 是cores包中types.go中定义的结构体实例
// 输出：result是格式化后的字符串
//序列化成json文件
result, err := analyzer.SerializeToJson(node)
//序列化成yaml文件
result, err := analyzer.SerializeToYaml(node)
```



## 2. 格式验证+反序列化（deserialize.go)

格式验证只提供了对json文件的格式验证,调用的示例如下：

```go
// 调用函数Deserialize进行格式验证和反序列化；
// 输入参数：json字符串 ， cores包中types.go中定义的结构体
// 返回值： node结构体
node, err = analyzer.Deserialize(str ，cores.strcutName{})
```
