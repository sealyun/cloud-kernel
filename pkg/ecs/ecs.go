package ecs

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/sealyun/cloud-kernel/pkg/exit"
	"github.com/sealyun/cloud-kernel/pkg/logger"
	"github.com/sealyun/cloud-kernel/pkg/vars"
	"strconv"
	"sync"
)

var once sync.Once
var hkCli *ecs.Client

func getClient() *ecs.Client {
	once.Do(func() {
		var err error
		vars.LoadVars()
		hkCli, err = ecs.NewClientWithAccessKey("", vars.AkId, vars.AkSK)
		if err != nil {
			exit.ProcessError(err)
		}
	})
	return hkCli
}

func getRegion() map[string]*ecs.RunInstancesRequest {
	region := make(map[string]*ecs.RunInstancesRequest)
	hk := ecs.CreateRunInstancesRequest()
	hk.ImageId = "centos_7_04_64_20G_alibase_201701015.vhd"
	hk.InstanceType = "ecs.c5.xlarge"
	hk.InternetChargeType = "PayByTraffic"
	hk.InternetMaxBandwidthIn = "50"
	hk.InternetMaxBandwidthOut = "50"
	hk.InstanceChargeType = "PostPaid"
	hk.SpotStrategy = "SpotAsPriceGo"
	hk.RegionId = "cn-hongkong"
	hk.SecurityGroupId = "sg-j6cb45dolegxcb32b47w"
	hk.VSwitchId = "vsw-j6cvaap9o5a7et8uumqyx"
	hk.ZoneId = "cn-hongkong-c"
	hk.Password = vars.EcsPassword
	region["cn-hongkong"] = hk
	return region
}

func New(amount int, dryRun bool, region string, bandwidthOut bool) []string {
	if region == "" {
		region = "cn-hongkong"
	}
	client := getClient()
	// 创建请求并设置参数
	request := getRegion()[region]
	request.Amount = requests.Integer(strconv.Itoa(amount))
	request.ClientToken = utils.GetUUID()
	if !bandwidthOut {
		request.InternetMaxBandwidthOut = "0"
	}
	request.DryRun = requests.Boolean(strconv.FormatBool(dryRun))
	response, err := client.RunInstances(request)
	if err != nil {
		_ = exit.ProcessError(err)
		return nil
	}
	return response.InstanceIdSets.InstanceIdSet
}

func Delete(dryRun bool, instanceId []string, region string) error {
	client := getClient()
	if region == "" {
		region = "cn-hongkong"
	}
	// 创建请求并设置参数
	request := ecs.CreateDeleteInstancesRequest()
	request.DryRun = requests.Boolean(strconv.FormatBool(dryRun))
	request.Force = "true"
	request.RegionId = "cn-hongkong"
	request.InstanceId = &instanceId
	response, err := client.DeleteInstances(request)
	if err != nil {
		_ = exit.ProcessError(err)
		logger.Error("递归删除ecs")
		return Delete(dryRun, instanceId, region)
	}
	logger.Info("删除成功: %s", response.RequestId)
	return nil
}

func Describe(instanceId string, region string) (*ecs.DescribeInstanceAttributeResponse, error) {
	client := getClient()
	if region == "" {
		region = "cn-hongkong"
	}
	request := ecs.CreateDescribeInstanceAttributeRequest()
	request.RegionId = region
	request.InstanceId = instanceId
	return client.DescribeInstanceAttribute(request)
}
