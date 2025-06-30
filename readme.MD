# 1.fsa 根据proto生成对应文件的处理过程
## 1.1proto处理
### 删除fsa proto文件中含有_get的注释行，目的是让functionId与property绑定
![fsaProtoDelGetLine](./images/fsaProtoDelGetLine.png "Optional title")
修改如下位置，改为fsa proto的全路径

![fsaDealProtoScript](./images/fsaDealProtoScript.png "Optional title")


在linux环境下执行dealFsaProto.sh脚本，将输出结果复制到一个新的proto文件中，如myv178new.proto

在proto的最上面，添加如下几行
```protobuf
syntax = "proto2";
package  fsa;
option  go_package = "com.fsa";
```
## 1.2根据proto生成json文件
```shell
protoc --doc_out=. --doc_opt=json,FsaDataSource.json --proto_path=/Users/wulili/Desktop/portal/depend/ --proto_path=/Users/wulili/Desktop/  /Users/wulili/Desktop/myv178new.proto

```

生成的json文件放到tmp目录下

![fsaDataSource](./images/fsaDataSource.png "Optional title")
## 1.3根据proto生成.go文件
```shell
protoc --go_out=.  --proto_path=/Users/wulili/Desktop/portal/depend/ --proto_path=/Users/wulili/Desktop/  /Users/wulili/Desktop/myv178new.proto
```
生成的.go文件放到如下目录中

![fsaGoCodeDir](./images/fsaGoCodeDir.png "Optional title")

## 1.4生成protoMarshal.go 和protoUnmarshal.go的代码
调用如下接口，生成相关代码
![fsaProtoMarshalAndUnmarshalApi](./images/fsaProtoMarshalAndUnmarshalApi.png "Optional title")

生成的代码会默认放到如下目录中

![fsaProtoMarshalGoDir](./images/fsaProtoMarshalGoDir.png "Optional title")

# 2.fsa demo案例运行步骤

报文交互图


![NFSA 仿真报文](./images/NFSA 仿真报文.jpg "Optional title")
## 2.1MyModuleNfsa.proto内容如下，是手动修改的可用的，不需要执行dealFsaProto.sh脚本进行处理
```protobuf
syntax = "proto2";
package fas;
option  go_package = "com.nfsa.mymodule";


//Method 1001 method
message method_request_response {
optional string task = 1; //客户端请求任务
}
message method_response {
optional bool result = 1; //服务端执行结果
}

//Method 1002 method
message method_request {
optional string task = 1; //客户端请求任务
}

//Property 1003 property
message property_status {
optional bool isOK = 1; //某个属性的状态
}
message property_get {
optional bool isOK = 1; //某个属性的状态
}
message property_set {
optional bool isOK = 1; //某个属性的状态
}

//Event 1004 event
message event_event {
optional string info = 1; //需要发布的消息
}
//**********************
//Interface ID: 3 iNotification
//**********************

//Method 120 Subscribe
//--------------------------------------
message  Subscribe_request_response  {

	repeated	 uint32	 FunctionID = 1 ;
}

message  Subscribe_response  {

	repeated	 functionResults	 FunctionResults = 1 ;
}

//Method 121 RemoveSubscription
//--------------------------------------
message  RemoveSubscription_request_response  {

	repeated	 uint32	 FunctionID = 1 ;
}

message  RemoveSubscription_response  {

	repeated	 functionResults	 FunctionResults = 1 ;
}

message  functionResults  {

	optional	 uint32	 FunctionID = 1 ;
	optional	 bool	 Successful = 2 ;
}


//**********************
//Interface ID: 4 iStateOfHealth
//**********************

//Property 100 Heartbeat
//--------------------------------------
message  Heartbeat_status  {

	optional	 uint32	 Heartbeat = 1 ;
}

//Event 111 RemoveService
//--------------------------------------
//message  RemoveService_event  {
//
//No parameters, so GPB not necessary.
//}

```

## 2.2根据proto生成json

```shell
 protoc --doc_out=. --doc_opt=json,FsaDataSource.json --proto_path=/Users/wulili/Desktop/portal/depend/ --proto_path=/Users/wulili/Desktop/  /Users/wulili/Desktop/MyModuleNfsa.proto
```
生成的FsaDataSource.json放到tmp目录下


## 2.3根据proto生成go
生成的go文件放到如下位置

![fsaMyModuleGoCode](./images/fsaMyModuleGoCode.png"Optional title")

## 2.4调用接口生成protoMarshal.go和protoUnmarshal.go
调用如下接口，生成相关代码
![fsaProtoMarshalAndUnmarshalApi](./images/fsaProtoMarshalAndUnmarshalApi.png "Optional title")

生成的代码会默认放到如下目录中

![fsaProtoMarshalGoDir](./images/fsaProtoMarshalGoDir.png "Optional title")

## 2.5FsaInterface.json做如下配置
![fsaInterface](./images/fsaInterfaceJson.png "Optional title")

```json
[{
  "service_name": "ParkingHMI",
  "service_id": 1006,
  "instance_id": 1,
  "transport_protocol": "tcp",
  "servers": [{
    "ecu": "ecu1",
    "ip": "127.0.0.1",
    "port": 9010,
    "mac": "mac"
  }],
  "clients": [{
    "ecu": "ecu2",
    "ip": "172.16.4.10ew",
    "port": 54321,
    "mac": "mac"
  }]
}]

```
## 2.6nfsa代码调整
修改ip为10.0.2.2   port为9010，和FsaInterface.json中server端的端口保持一致
![nfsaCodeIPAndPort](./images/nfsaCodeIPAndPort.png "Optional title")
## 2.7动nfsa

![startNfsa](./images/startNfsa.png "Optional title")

## 2.8设置method和property的默认response


![fsaSetDefaultResponse](./images/fsaSetDefaultResponse.png "Optional title")
## 2.9启动andriod nfsa demo
从日志中可看到我们设置的默认回复值

![fsaAndriodDemoGetDefaultResponse](./images/fsaAndriodDemoGetDefaultResponse.png "Optional title")

## 2.10发送event
测试时发现，值不变，也会触发android端日志打印

![fsaSendEvent](./images/fsaSendEvent.png "Optional title")

andriod中的fsa client接收到event
![fsaClientReceiveEvent](./images/fsaClientReceiveEvent.png "Optional title")

## 2.11发送proerty

**测试时发现，值不变，不会触发android端日志打印**

![fsaSendEvent](./images/sendFsaProperty.png "Optional title")

andriod中的fsa client接收到property

![fsaAndriodReceiveProperty](./images/fsaAndriodReceiveProperty.png "Optional title")

## 2.12订阅method和propretty
![subscribeFsa](./images/subscribeFsa.png "Optional title")

重启nfsa andriod demo,订阅页面会有数据

![receiveFsaSubscribe](./images/receiveFsaSubscribe.png "Optional title")
