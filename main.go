package main

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"simulator_backend/api/controller"
	"simulator_backend/api/service"
	"simulator_backend/bdd"
	"simulator_backend/initialization"
	"simulator_backend/mockServices"
	"simulator_backend/util/adb"
	"simulator_backend/util/can"
	"simulator_backend/util/file"
	logutil "simulator_backend/util/log"
	"simulator_backend/websocket"
)

// resourceRouter 静态资源配置
func resourceRouter(engine *gin.Engine) {
	html := controller.NewHtmlHandler()
	group := engine.Group("/ui")
	{
		group.GET("", html.Index)
	}
	// 解决刷新404问题
	engine.NoRoute(html.RedirectIndex)
}
func InitResource(engine *gin.Engine) *gin.Engine {
	engine.StaticFS("/js", http.FS(initialization.NewResource()))
	engine.StaticFS("/css", http.FS(initialization.NewResource()))
	engine.StaticFS("/static", http.FS(initialization.NewResource()))
	return engine
}
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		go func() {
			requestUrl := c.Request.URL.Path
			//log.Println("requestUrl-------:" + requestUrl)
			auditService := service.NewAuditService()
			err := auditService.CreateAudit(requestUrl)
			if err != nil {
				log.Println(fmt.Sprintf("CreateAudit err,err:%s", err.Error()))
			}
		}()

	}
}
func main() {

	adb.InitDeviceFlag()

	hasEmulator := flag.Bool("hasEmulator", true, "hasEmulator:本地是否有启动对应的andriod模拟器，默认为true")
	startCan := flag.Bool("startCan", true, "startCan:接收pythonCan的数据，默认为true")

	startBddServer := flag.Bool("startBddServer", true, "StartBddServer:启动bddserver")
	companyDomain := flag.Bool("companyDomain", false, "companyDomain:是否为域内网络，默认为false，链接的是域外网")
	hasNetwork := flag.Bool("hasNetwork", true, "hasNetwork:电脑是否联网，默认为true，联网状态")

	flag.Parse()
	if *startBddServer {
		go func() {
			bddServer := bdd.NewBddServer()
			err := bddServer.StartBddServer()
			if err != nil {
				log.Println(fmt.Sprintf("start bdd err,err:%v", err.Error()))
				os.Exit(1)
			}
		}()
	}
	if *hasEmulator {
		err := adb.InitSimulationProxy()
		if err != nil {
			log.Println(fmt.Sprintf("InitSimulationProxy err,err:%v", err.Error()))
			os.Exit(1)
		}
	}

	err := logutil.InitLog()
	if err != nil {
		log.Println(fmt.Sprintf("init log err,err:%v", err.Error()))
		os.Exit(1)
	}
	//初始化someipDatasource
	err = file.LoadSomeIpData()
	if err != nil {
		log.Println(fmt.Sprintf("LoadSomeIpData err,err:%v", err.Error()))
		os.Exit(1)
	}
	//初始化fsaDataSource
	err = file.LoadFsaData()
	if err != nil {
		log.Println(fmt.Sprintf("LoadFsaData err,err:%v", err.Error()))
		os.Exit(1)
	}

	err = file.RemoveFile(file.MockServiceStatusFile)
	if err != nil {
		log.Println(err.Error())
	}
	//// 接收proxy端的rpc请求
	//controller.MockServiceFunction()
	engine := gin.Default()
	//处理跨域问题、
	engine.Use(cors.Default())
	resourceRouter(engine)
	engine = InitResource(engine)
	// websocket 处理发布订阅的逻辑
	ws := websocket.NewWebsocketHandler()
	wsGroup := engine.Group("/ws")
	if *hasNetwork {
		//有网，才开启审计日志的记录
		wsGroup.Use(AuditMiddleware())
	}

	{
		wsGroup.GET("/publish", ws.Publish)
		wsGroup.GET("/subscribe", ws.Subscribe)
		wsGroup.GET("/rpcCall", ws.RpcCall)
		wsGroup.GET("/startMockService", ws.StartMockService)
		wsGroup.GET("/subscribeSomeIpSignal", ws.SubscribeSomeIpSignal)
		wsGroup.GET("/subscribeCanSignal", ws.SubscribeCanSignal)
		wsGroup.GET("/sendCanSignal", ws.SendCanSignal)
		wsGroup.GET("/subscribeRpcParam", ws.SubscribeRpcParam)
		wsGroup.GET("/adasRpcCall", ws.AdasRpcCall)
		wsGroup.GET("/adasSendTopics", ws.AdasSendTopics)
		wsGroup.GET("/simulateClientAdasRpcCall", ws.SimulateClientAdasRpcCall)
		wsGroup.GET("/simulateClientSubTopics", ws.SimulateClientSubTopics)

		wsGroup.GET("/simulateSomeIpClientSendAndSubRes", ws.SimulateSomeIpClientSendAndSubRes)

		wsGroup.GET("/subFsa", ws.SubscribeFsa)
		wsGroup.GET("/parseNetPacket", ws.ParseNetPacket)
		//d2c
		wsGroup.GET("/dtcSubscribeRpcParam", ws.DtcSubscribeRpcParam)
		wsGroup.GET("/dealSubscribeFromCloud", ws.DealSubscribeFromCloud)
		//c2d
		wsGroup.GET("/ctdSubscribeToVehicle", ws.CtdSubscribeToVehicle)

	}

	//订阅的接口
	pub := controller.NewPubHandler()
	//rpc接口
	rpc := controller.NewRpcHandler()
	mockService := controller.NewMockServiceHandler()
	someIpService := controller.NewSomeIpHandler()
	autoTest := controller.NewAutotestHandler()
	canService := controller.NewCanHandler()
	autoTestPlan := controller.NewAutoTestPlanHandler()
	autoTestScenario := controller.NewAutoTestScenarioHandler()
	autoTestStep := controller.NewAutoTestStepHandler()
	goPacketHandler := controller.NewGoPacketHandler()
	cloudLogHandler := controller.NewCloudLogHandler()
	fsaHandler := controller.NewFsaHandler()
	dtcHandler := controller.NewDtcHandler()
	ctdHandler := controller.NewCtdHandler()

	gin.SetMode(gin.ReleaseMode)
	apiGroup := engine.Group("/api")
	if *hasNetwork {
		apiGroup.Use(AuditMiddleware())
	}

	{
		apiGroup.GET("/topicMessages", pub.GetTopicMessages)
		apiGroup.POST("/getTopicTree", pub.GetTopicTree)

		apiGroup.POST("/topicMessageFields", pub.GetTopicMessageFields)

		apiGroup.GET("/methods", rpc.GetRpcMethods)

		apiGroup.POST("/getMethodTree", rpc.GetMethodTree)
		apiGroup.POST("/methodParam", rpc.GetRpcMessageFields)
		apiGroup.GET("/mockServices", mockService.GetMockServices)
		apiGroup.GET("/getMockServiceStatus", mockService.GetMockServiceStatus)
		apiGroup.POST("/uploadSomeIpExcel", someIpService.UploadSomeIpExcel)
		apiGroup.GET("/getSomeIpTableData", someIpService.GetSomeIpTableData)
		apiGroup.POST("/sendSomeIpSignal", someIpService.SendSomeIpSignal)
		//apiGroup.POST("/autoTestTopic", autoTest.AutoTestTopic)
		//apiGroup.POST("/autoTestRpc", autoTest.AutoTestRpc)
		apiGroup.POST("/autoTest", autoTest.AutoTest)
		apiGroup.POST("/uploadCanExcel", canService.UploadCanExcel)
		apiGroup.GET("/getCanMessage", canService.GetCanMessage)
		apiGroup.POST("/sendCanMessage", canService.SendCanMessage)
		apiGroup.POST("/uploadTestPlan", autoTest.UploadTestPlan)
		apiGroup.POST("/executeTestPlan", autoTestPlan.ExecuteTestPlan)
		apiGroup.POST("/executeLastFailedTestPlan", autoTestPlan.ExecuteLastFailedTestPlan)
		apiGroup.POST("/executePartialTestPlan", autoTestPlan.ExecutePartialTestPlan)
		apiGroup.POST("/downloadTestPlan", autoTest.DownloadTestPlan)
		apiGroup.POST("/getAllTestPlanResult", autoTest.GetAllTestPlanResult)
		apiGroup.POST("/getTestPlanResultDetail", autoTest.GetTestPlanResultDetail)
		apiGroup.POST("/downloadAllTestPlanResult", autoTest.DownloadAllTestPlanResult)
		apiGroup.POST("/addTestPlan", autoTestPlan.AddTestPlan)
		apiGroup.POST("/getTestPlanName", autoTestPlan.GetTestPlanName)
		apiGroup.POST("/getAllTestPlan", autoTestPlan.GetAllTestPlan)
		apiGroup.POST("/deleteTestPlan", autoTestPlan.DeleteTestPlan)
		apiGroup.POST("/editTestPlanName", autoTestPlan.EditTestPlanName)
		apiGroup.POST("/createScenario", autoTestScenario.CreateScenario)
		apiGroup.POST("/getScenarioListWithDetail", autoTestScenario.GetScenarioListWithDetail)
		apiGroup.POST("/deleteScenario", autoTestScenario.DeleteScenario)
		apiGroup.POST("/moveScenario", autoTestScenario.MoveScenario)
		apiGroup.POST("/editScenario", autoTestScenario.EditScenario)
		apiGroup.POST("/copyScenario", autoTestScenario.CopyScenario)
		apiGroup.POST("/createTestCase", autoTestScenario.CreateTestCase)
		apiGroup.POST("/editTestCase", autoTestScenario.EditTestCase)
		apiGroup.POST("/copyTestCase", autoTestScenario.CopyTestCase)
		apiGroup.POST("/deleteTestCase", autoTestScenario.DeleteTestCase)
		apiGroup.POST("/rpcExecCompareExpectAndActual", autoTestPlan.RpcExecCompareExpectAndActual)
		apiGroup.POST("/topicExecCompareExpectAndActual", autoTestPlan.TopicExecCompareExpectAndActual)
		apiGroup.POST("/moveTestCase", autoTestScenario.MoveTestCase)
		apiGroup.POST("/addTestCaseStep", autoTestStep.AddTestCaseStep)
		apiGroup.POST("/moveTestCaseStep", autoTestStep.MoveTestCaseStep)
		apiGroup.POST("/deleteTestCaseStep", autoTestStep.DeleteTestCaseStep)
		apiGroup.POST("/saveTestCaseStepConfig", autoTestStep.SaveTestCaseStepConfig)
		apiGroup.POST("/exportTpResults", autoTest.ExportTpResults)
		apiGroup.POST("/exportTps", autoTest.ExportTps)

		apiGroup.POST("/getAdasMethods", someIpService.GetAdasMethods)
		apiGroup.POST("/getAdasTopics", someIpService.GetAdasTopics)
		apiGroup.POST("/getSomeIpServiceDefinition", someIpService.GetSomeIpServiceDefinition)

		apiGroup.POST("/startSomeIp", someIpService.StartManySomeIp)
		apiGroup.POST("/getCurrentSomeIpStarted", someIpService.GetCurrentSomeIpStarted)
		apiGroup.POST("/stopSomeIpServer", someIpService.StopSomeIpServer)
		apiGroup.POST("/getSomeIpModules", someIpService.GetSomeIpModules)
		apiGroup.POST("/getModuleInterfaces", someIpService.GetModuleInterfaces)

		apiGroup.POST("/findAllDevice", goPacketHandler.FindAllDevice)
		apiGroup.POST("/parseNetPacket", goPacketHandler.ParseNetPacket)
		apiGroup.POST("/getNetPacket", goPacketHandler.GetNetPacket)

		//someip报文回放
		apiGroup.GET("/replaySomeIp", goPacketHandler.ReplaySomeIp)

		apiGroup.POST("/fsaMessageFields", fsaHandler.GetFsaMessageFields)
		apiGroup.POST("/fsaProtoMarshalAndUnmarshal", fsaHandler.GenerateFsaProtoMarshalAndUnmarshal)
		apiGroup.POST("/startFsa", fsaHandler.StartFsa)
		apiGroup.POST("/stopFsa", fsaHandler.StopFsa)
		apiGroup.POST("/getFsaFunctions", fsaHandler.GetFsaFunctions)
		apiGroup.POST("/setFsaDefaultResponse", fsaHandler.SetFsaDefaultResponse)
		apiGroup.POST("/getFsaInterfaces", fsaHandler.GetFsaInterfaces)
		apiGroup.POST("/getFsaStarted", fsaHandler.GetFsaStarted)
		apiGroup.POST("/sendFsa", fsaHandler.SendFsa)
		apiGroup.POST("/getFsaDefaultResponse", fsaHandler.GetFsaDefaultResponse)
		apiGroup.GET("/sendFsaJsonData", fsaHandler.SendFsaJsonData)

		apiGroup.GET("/getCloudRpcLog", cloudLogHandler.GetCloudRpcLog)
		apiGroup.GET("/publishCycle", pub.PublishCycle)
		//dtc
		apiGroup.POST("/generateDTCCloudEvent", dtcHandler.GenerateDTCCloudEvent)
		apiGroup.POST("/sendDTC", dtcHandler.SendDTC)
		apiGroup.POST("/generateDTCRpcCloudEvent", dtcHandler.GenerateDTCRpcCloudEvent)
		apiGroup.POST("/vehicleOnline", dtcHandler.VehicleOnline)
		apiGroup.POST("/vehicleOffline", dtcHandler.VehicleOffline)
		apiGroup.POST("/getOnlineVehicles", dtcHandler.GetOnlineVehicles)
		apiGroup.POST("/selectVehicle", dtcHandler.SelectVehicle)
		apiGroup.POST("/cancelSelectVehicle", dtcHandler.CancelSelectVehicle)
		apiGroup.POST("/getSubscriptionData", dtcHandler.GetSubscriptionData)
		apiGroup.POST("/saveRpcResponseSettings", dtcHandler.SaveRpcResponseSettings)
		apiGroup.POST("/getRpcResponseSettings", dtcHandler.GetRpcResponseSettings)

		//ctd
		apiGroup.POST("/getCtdMqinfo", ctdHandler.GetCtdMqinfo)
		apiGroup.POST("/generateCTDCloudEvent", ctdHandler.GenerateCTDCloudEvent)
		apiGroup.POST("/sendManyCTDRpc", ctdHandler.SendManyCTDRpc)
		apiGroup.POST("/getManyCTDRpc", ctdHandler.GetManyCTDRpc)
		apiGroup.POST("/generateCTDSubscribeCloudEvent", ctdHandler.GenerateCTDSubscribeCloudEvent)
		apiGroup.POST("/parseCeBase64", ctdHandler.ParseCeBase64)

	}

	//pprof 监控
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	if *hasNetwork {
		if *companyDomain {
			//域内网
			err := initialization.InitCompanyDomainMysql()
			if err != nil {
				log.Println(fmt.Sprintf("InitCompanyDomainMysql err,err:%v", err.Error()))
				os.Exit(1)
			}

		} else {
			//域外网
			err := initialization.InitNotCompanyDomainMysql()
			if err != nil {
				log.Println(fmt.Sprintf("InitNotCompanyDomainMysql err,err:%v", err.Error()))
				os.Exit(1)
			}
			//初始化mqtt，为了与proxy相连
			err = initialization.InitMqttClient()
			if err != nil {
				log.Println(fmt.Sprintf("init mqtt err,err:%v", err.Error()))
				os.Exit(1)
			}

			go func() {
				mockServiceServer := mockServices.NewMockServiceServer()
				mockServiceServer.MockServiceCallStart()
			}()

		}

	}

	if *startCan {
		go func() {
			canServer := can.NewCanServer()
			err = canServer.StartCanServer()
			if err != nil {
				log.Println(fmt.Sprintf("启动can server报错，报错信息为：%s", err.Error()))
			}
		}()
	}

	err = engine.Run(":8877")
	if err != nil {
		panic(fmt.Sprintf("run server err,err is:%s", err.Error()))
	}

	http.HandleFunc("/getAdasMethod", file.HttpGetAdasMethods)
	log.Fatal(http.ListenAndServe(":9090", nil))

}
