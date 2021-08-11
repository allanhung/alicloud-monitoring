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
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"

	"github.com/allanhung/alicloud-monitoring/pkg/alicloud"
	"github.com/allanhung/alicloud-monitoring/pkg/joblock"
	"github.com/allanhung/alicloud-monitoring/pkg/log"
	"github.com/allanhung/alicloud-monitoring/pkg/monitor"
)

var spotPriceQueryFlags = alicloud.QuerySpotPriceFlags{}

// spotPriceCmd represents the ecs command
var spotPriceCmd = &cobra.Command{
	Use:   "spotprice",
	Short: "Query spot price for kubernetes worker.",
	Long: `This tool will query spot instance price for kubernetes worker.

example:
  alicloud-monitoring spotprice --logfile /tmp/ecs_update.log --loglevel debug
  alicloud-monitoring spotprice --cron '0 */5 * * * *'`,
	Run: func(cmd *cobra.Command, args []string) {

		pm := monitor.NewSpotMonitor()
		var c *cron.Cron
		jobLock := joblock.JobLock{}

		pm.SpotPriceWatchdog.With(prometheus.Labels{"name": cmd.Use}).Set(1)
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

		if spotPriceQueryFlags.Cron == "" {
			err := querySpotPrice(jobLock, aliClient, pm)
			if err != nil {
				log.Logger.Errorf("%v", err)
				os.Exit(1)
			}
		} else {
			c = cron.New(cron.WithSeconds())
			c.AddFunc(spotPriceQueryFlags.Cron, func() { querySpotPrice(jobLock, aliClient, pm) })
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

func stringInList(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func queryEcsTypesInk8s(job joblock.JobLock, aliClient *alicloud.AliClient) ([]string, error) {
	instanceTypes := []string{}
	var ecsQueryFlags = alicloud.QueryEcsFlags{}
	ecsQueryFlags.ReName = append(ecsQueryFlags.ReName, "worker-k8s.*")
	ecsQueryFlags.PageSize = spotPriceQueryFlags.PageSize

	queryList, err := alicloud.QueryECS(aliClient, ecsQueryFlags)
	if err != nil {
		return nil, err
	}

	for _, ecsInstance := range queryList {
		if !(stringInList(ecsInstance.InstanceType, instanceTypes)) && strings.HasPrefix(ecsInstance.SpotStrategy, "Spot") {
			instanceTypes = append(instanceTypes, ecsInstance.InstanceType)
		}
	}
	return instanceTypes, nil
}

func querySpotPrice(job joblock.JobLock, aliClient *alicloud.AliClient, pm *monitor.SpotMonitor) error {
	if job.IsRunning {
		return fmt.Errorf("job is still running: %s", job.Kind)
	} else {
		job.SetRun("Checking spot price")
	}
	log.Logger.Infof("Running job: %s", job.Kind)

	instanceTypes, err := queryEcsTypesInk8s(job, aliClient)
	if err != nil {
		return err
	}

	for _, instanceType := range instanceTypes {
		zoneList := []string{}
		spotPrices, err := alicloud.QuerySpotPrice(aliClient, instanceType)
		if err != nil {
			return err
		}
		for _, spotPrice := range spotPrices {
			if !stringInList(spotPrice.ZoneId, zoneList) {
				pm.SpotPrice.With(prometheus.Labels{"zoneid": spotPrice.ZoneId, "type": instanceType}).Set(spotPrice.SpotPrice)
				pm.ListPrice.With(prometheus.Labels{"zoneid": spotPrice.ZoneId, "type": instanceType}).Set(spotPrice.OriginPrice)
				zoneList = append(zoneList, spotPrice.ZoneId)
				log.Logger.Debugf("Instance Type: %s, Zone: %s, Spot Price: %v, List Price: %v", instanceType, spotPrice.ZoneId, spotPrice.SpotPrice, spotPrice.OriginPrice)
			}
		}
	}
	job.DoneRun()
	log.Logger.Infof("Job Completed.")
	return nil
}

func init() {
	rootCmd.AddCommand(spotPriceCmd)
	f := spotPriceCmd.Flags()
	f.IntVarP(&spotPriceQueryFlags.PageSize, "pagesize", "s", 10, "alicloud api pagesize")
	f.StringVarP(&spotPriceQueryFlags.Region, "region", "r", "us-east-1", "query region")
	f.StringVarP(&spotPriceQueryFlags.Cron, "cron", "c", "", "cron scheduler")
}
