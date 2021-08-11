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
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"

	"github.com/allanhung/alicloud-monitoring/pkg/alicloud"
	"github.com/allanhung/alicloud-monitoring/pkg/joblock"
	"github.com/allanhung/alicloud-monitoring/pkg/log"
	"github.com/allanhung/alicloud-monitoring/pkg/monitor"
)

var ecsCmdFlags = alicloud.QueryEcsFlags{}

// ecsCmd represents the ecs command
var ecsCmd = &cobra.Command{
	Use:   "ecs",
	Short: "ECS tag update",
	Long: `This tool will update ecs tag for ECS running on Alicloud. 

example:
  alicloud-monitoring ecs --regname 'worker-k8s.*' --logfile /tmp/ecs_update.log --loglevel debug
  alicloud-monitoring ecs --regname 'worker-k8s.*' --notagk Environment --cron '* * * * * *'`,
	Run: func(cmd *cobra.Command, args []string) {
		pm := monitor.NewTagsMonitor()
		var c *cron.Cron
		jobLock := joblock.JobLock{}
		instanceList := map[string]ecs.Instance{}

		cfg := &alicloud.AliCloudConfig{}
		err := cfg.GetCloudConfig()
		if err != nil {
			log.Logger.Errorf("failed to getCloudConfigFromStsToken: %v", err)
			os.Exit(1)
		}

		aliClient, err := alicloud.NewAliClient(cfg)
		if err != nil {
			log.Logger.Errorf("failed to create aliClient: %v", err)
			os.Exit(1)
		}

		if ecsCmdFlags.Cron == "" {
			err := queryECStag(jobLock, aliClient, ecsCmdFlags, pm, instanceList)
			if err != nil {
				log.Logger.Errorf("%v", err)
				os.Exit(1)
			}
			log.Logger.Debugf("%v", instanceList)
		} else {
			c = cron.New(cron.WithSeconds())
			c.AddFunc(ecsCmdFlags.Cron, func() { queryECStag(jobLock, aliClient, ecsCmdFlags, pm, instanceList) })
			fmt.Printf("start: %v", time.Now())
			c.Start()
		}
		err = monitor.PrometheusBoot()
		c.Stop()
		if err != nil {
			log.Logger.Errorf("failed to ListenAndServe: %v", err)
			os.Exit(1)
		}
	}}

func queryECStag(job joblock.JobLock, aliClient *alicloud.AliClient, queryFlags alicloud.QueryEcsFlags, pm *monitor.TagsMonitor, instanceList map[string]ecs.Instance) error {

	if job.IsRunning {
		return fmt.Errorf("job is still running: %s", job.Kind)
	} else {
		job.SetRun("Query")
	}

	log.Logger.Infof("Running job: %s", job.Kind)
	vpcMap, err := getVPCInfo(aliClient, ecsCmdFlags.PageSize)
	if err != nil {
		return fmt.Errorf("failed to get VPC information: %v", err)
	}

	for k, ecsInstance := range instanceList {
		pm.NoEnvTag.With(prometheus.Labels{"id": ecsInstance.InstanceId, "vpc": vpcMap[ecsInstance.VpcAttributes.VpcId], "name": ecsInstance.InstanceName}).Set(0)
		delete(instanceList, k)
	}

	queryList, err := alicloud.QueryECS(aliClient, queryFlags)
	if err != nil {
		return err
	}

	for _, ecsInstance := range queryList {
		instanceList[ecsInstance.InstanceId] = ecsInstance
		pm.NoEnvTag.With(prometheus.Labels{"id": ecsInstance.InstanceId, "vpc": vpcMap[ecsInstance.VpcAttributes.VpcId], "name": ecsInstance.InstanceName}).Set(1)
		log.Logger.Infof("instance: %s (%s) is in environment %s", ecsInstance.InstanceId, ecsInstance.InstanceName, vpcMap[ecsInstance.VpcAttributes.VpcId])
	}
	job.DoneRun()
	return nil
}

func getVPCInfo(aliClient *alicloud.AliClient, pageSize int) (map[string]string, error) {
	vpcMap := map[string]string{}
	allVpcs, err := alicloud.QueryVpc(aliClient, pageSize)
	if err != nil {
		return vpcMap, err
	}

	for _, Vpc := range allVpcs {
		switch Vpc.VpcName {
		case "dev":
			vpcMap[Vpc.VpcId] = "develop"
		default:
			vpcMap[Vpc.VpcId] = Vpc.VpcName
		}
	}

	return vpcMap, nil
}

func init() {
	rootCmd.AddCommand(ecsCmd)
	f := ecsCmd.Flags()
	f.StringVarP(&ecsCmdFlags.InstanceName, "instancename", "n", "", "filter by instance name")
	f.IntVarP(&ecsCmdFlags.PageSize, "pagesize", "s", 10, "alicloud api pagesize")
	f.VarP(&ecsCmdFlags.Tag, "tag", "t", "filter by ecs instance tag example: cluster=prod (can specify multiple)")
	f.VarP(&ecsCmdFlags.ReName, "re", "", "filter by ecs instance name with regular expression  example: ecs.* (can specify multiple, will use or operator)")
	f.VarP(&ecsCmdFlags.NoTagKey, "notagk", "", "filter by ecs instance tag key not contain keyword with regular expression example: acs:autoscaling.* (can specify multiple)")
	f.VarP(&ecsCmdFlags.NoTagValue, "notagv", "", "filter by ecs instance tag value not contain keyword with regular expression example: autoScale (can specify multiple)")
	f.StringVarP(&ecsCmdFlags.Cron, "cron", "c", "", "cron scheduler")
}
