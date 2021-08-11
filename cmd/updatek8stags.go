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

var updateK8sTagsCmdFlags = alicloud.QueryEcsFlags{}

// updateK8sTagsCmd represents the ecs command
var updateK8sTagsCmd = &cobra.Command{
	Use:   "updatek8stags",
	Short: "ECS tag update for kubernetes worker",
	Long: `This tool will update ecs tag for kubernetes worker running on Alicloud. 

example:
  alicloud-monitoring updatek8stags --logfile /tmp/ecs_update.log --loglevel debug
  alicloud-monitoring updatek8stags --cron '0 * * * * *'`,
	Run: func(cmd *cobra.Command, args []string) {

		pm := monitor.NewTagsMonitor()
		var c *cron.Cron
		jobLock := joblock.JobLock{}
		instanceList := map[string]ecs.Instance{}

		pm.NoEnvTagWatchdog.With(prometheus.Labels{"name": cmd.Use}).Set(1)
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

		if updateK8sTagsCmdFlags.Cron == "" {
			err := addk8sTags(jobLock, aliClient, pm, instanceList)
			if err != nil {
				log.Logger.Errorf("%v", err)
				os.Exit(1)
			}
			log.Logger.Debugf("%v", instanceList)
		} else {
			c = cron.New(cron.WithSeconds())
			c.AddFunc(updateK8sTagsCmdFlags.Cron, func() { addk8sTags(jobLock, aliClient, pm, instanceList) })
			fmt.Printf("start: %v", time.Now())
			c.Start()
		}
		err = monitor.PrometheusBoot()
		c.Stop()
		if err != nil {
			log.Logger.Errorf("failed to ListenAndServe: %v", err)
			os.Exit(1)
		}
	},
}

func addk8sTags(job joblock.JobLock, aliClient *alicloud.AliClient, pm *monitor.TagsMonitor, instanceList map[string]ecs.Instance) error {
	queryECStag(job, aliClient, updateK8sTagsCmdFlags, pm, instanceList)
	if job.IsRunning {
		return fmt.Errorf("job is still running: %s", job.Kind)
	} else {
		job.SetRun("Update")
	}

	log.Logger.Infof("Running job: %s", job.Kind)
	vpcMap, err := getVPCInfo(aliClient, ecsCmdFlags.PageSize)
	if err != nil {
		return fmt.Errorf("failed to get VPC information: %v", err)
	}
	for k, v := range instanceList {
		if (updateK8sTagsCmdFlags.InstanceId == "") || (updateK8sTagsCmdFlags.InstanceId != "" && k == updateK8sTagsCmdFlags.InstanceId) {
			k8sTag := []ecs.AddTagsTag{
				{
					Key:   "Environment",
					Value: vpcMap[v.VpcAttributes.VpcId],
				},
				{
					Key:   "role",
					Value: "worker",
				},
				{
					Key:   "stack",
					Value: "kubernetes",
				},
			}
			log.Logger.Infof("InstanceId：%s start update", k)
			err := alicloud.AddInstanceTags(aliClient, v, k8sTag)
			if err != nil {
				log.Logger.Errorf("InstanceId：%s failed to add tag. error: %v", k, err)
			}
			pm.NoEnvTag.With(prometheus.Labels{"id": k, "vpc": vpcMap[v.VpcAttributes.VpcId], "name": v.InstanceName}).Set(0)
		} else {
			log.Logger.Debugf("InstanceId：%s != %s will not update", k, updateK8sTagsCmdFlags.InstanceId)
		}
	}
	job.DoneRun()
	return nil
}

func init() {
	rootCmd.AddCommand(updateK8sTagsCmd)
	f := updateK8sTagsCmd.Flags()
	updateK8sTagsCmdFlags.NoTagKey = append(updateK8sTagsCmdFlags.NoTagKey, "Environment")
	updateK8sTagsCmdFlags.ReName = append(updateK8sTagsCmdFlags.ReName, "worker-k8s.*")
	f.StringVarP(&updateK8sTagsCmdFlags.InstanceId, "instanceid", "i", "", "filter by instance id")
	f.IntVarP(&updateK8sTagsCmdFlags.PageSize, "pagesize", "s", 10, "alicloud api pagesize")
	f.StringVarP(&updateK8sTagsCmdFlags.Cron, "cron", "c", "", "cron scheduler")
}
