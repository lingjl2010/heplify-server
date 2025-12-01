## 概述

这里实现了项目自己的自定义插件

目前包含了三项检查

1. httpctx 检查不能将非gin.Context/tgo.Context传给context.Context参数
2. executable 检查main文件必须引用内部的executable包
3. sqlctx 检查是否存在Context版本的函数并提示使用Context版本
4. protobuf 检查是否直接使用了proto.Unmarshal

## 使用方式

1. 作为golangci-lint的插件使用，优点是融入lint体系，支持nolint注释等。缺点是项目需要和golangci-lint的运行时版本绑定，目前采用的是v1.55.2的版本。参见plugin目录

2. 作为独立命令使用，直接运行即可，参见./main.go，优点是运行方便。
