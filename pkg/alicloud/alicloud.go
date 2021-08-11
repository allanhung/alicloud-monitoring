/*
Copyright © 2019 Allan Hung <hung.allan@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package alicloud

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/denverdino/aliyungo/metadata"
	"github.com/allanhung/alicloud-monitoring/pkg/log"
)

type AliCloudConfig struct {
	RegionID        string    `json:"regionId" yaml:"regionId"`
	AccessKeyID     string    `json:"accessKeyId" yaml:"accessKeyId"`
	AccessKeySecret string    `json:"accessKeySecret" yaml:"accessKeySecret"`
	VPCID           string    `json:"vpcId" yaml:"vpcId"`
	RoleName        string    `json:"-" yaml:"-"` // For ECS RAM role only
	StsToken        string    `json:"-" yaml:"-"`
	ExpireTime      time.Time `json:"-" yaml:"-"`
}

type AliClient struct {
	RegionID   string
	EcsClient  ecs.Client
	clientLock sync.RWMutex
	nextExpire time.Time
}

func (a *AliCloudConfig) GetCloudConfig() error {
	roleName := ""
	if os.Getenv("ALICLOUD_REGION") == "" ||
		os.Getenv("ALICLOUD_ACCESS_KEY") == "" ||
		os.Getenv("ALICLOUD_SECRET_KEY") == "" {
		httpClient := &http.Client{
			Timeout: 3 * time.Second,
		}
		// Load config from Metadata Service
		m := metadata.NewMetaData(httpClient)
		roleName, err := m.RoleName()
		if err != nil {
			return fmt.Errorf("failed to get role name from Metadata Service: %v", err)
		}
		vpcID, err := m.VpcID()
		if err != nil {
			return fmt.Errorf("failed to get VPC ID from Metadata Service: %v", err)
		}
		regionID, err := m.Region()
		if err != nil {
			return fmt.Errorf("failed to get Region ID from Metadata Service: %v", err)
		}
		role, err := m.RamRoleToken(roleName)
		if err != nil {
			return fmt.Errorf("failed to get STS Token from Metadata Service: %v", err)
		}
		a.RegionID = regionID
		a.RoleName = roleName
		a.VPCID = vpcID
		a.AccessKeyID = role.AccessKeyId
		a.AccessKeySecret = role.AccessKeySecret
		a.StsToken = role.SecurityToken
		a.ExpireTime = role.Expiration
	} else {
		a.RegionID = os.Getenv("ALICLOUD_REGION")
		a.AccessKeyID = os.Getenv("ALICLOUD_ACCESS_KEY")
		a.AccessKeySecret = os.Getenv("ALICLOUD_SECRET_KEY")
		a.RoleName = roleName
	}
	return nil
}

func NewAliClient(cfg *AliCloudConfig) (*AliClient, error) {
	var err error
	var ecsClient *ecs.Client

	if cfg.RoleName == "" {
		ecsClient, err = ecs.NewClientWithAccessKey(
			cfg.RegionID,
			cfg.AccessKeyID,
			cfg.AccessKeySecret,
		)
	} else {
		ecsClient, err = ecs.NewClientWithStsToken(
			cfg.RegionID,
			cfg.AccessKeyID,
			cfg.AccessKeySecret,
			cfg.StsToken,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create alicloud client: %v", err)
	}
	aliClient := &AliClient{
		RegionID:  cfg.RegionID,
		EcsClient: *ecsClient,
	}
	if cfg.RoleName != "" {
		aliClient.setNextExpire(cfg.ExpireTime)
		go aliClient.refreshStsToken(cfg, 1*time.Second)
	}
	return aliClient, nil
}

func (p *AliClient) setNextExpire(expireTime time.Time) {
	p.clientLock.Lock()
	defer p.clientLock.Unlock()
	p.nextExpire = expireTime
}

func (p *AliClient) refreshStsToken(cfg *AliCloudConfig, sleepTime time.Duration) {
	for {
		time.Sleep(sleepTime)
		now := time.Now()
		utcLocation, err := time.LoadLocation("")
		if err != nil {
			log.Logger.Errorf("Get utc time error %v", err)
			continue
		}
		nowTime := now.In(utcLocation)
		p.clientLock.RLock()
		sleepTime = p.nextExpire.Sub(nowTime)
		p.clientLock.RUnlock()
		log.Logger.Infof("Distance expiration time %v", sleepTime)
		if sleepTime < 10*time.Minute {
			sleepTime = time.Second * 1
		} else {
			sleepTime = 9 * time.Minute
			log.Logger.Info("Next fetch sts sleep interval : ", sleepTime.String())
			continue
		}
		err = cfg.GetCloudConfig()
		if err != nil {
			log.Logger.Errorf("Failed to refreshStsToken: %v", err)
			continue
		}
		var ecsClient *ecs.Client

		if cfg.RoleName == "" {
			ecsClient, err = ecs.NewClientWithAccessKey(
				cfg.RegionID,
				cfg.AccessKeyID,
				cfg.AccessKeySecret,
			)
		} else {
			ecsClient, err = ecs.NewClientWithStsToken(
				cfg.RegionID,
				cfg.AccessKeyID,
				cfg.AccessKeySecret,
				cfg.StsToken,
			)
		}

		log.Logger.Infof("Refresh client from sts token, next expire time %v", cfg.ExpireTime)
		p.clientLock.Lock()
		p.RegionID = cfg.RegionID
		p.EcsClient = *ecsClient
		p.nextExpire = cfg.ExpireTime
		p.clientLock.Unlock()
	}
}
