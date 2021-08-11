module github.com/allanhung/alicloud-monitoring

go 1.15

require (
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.435
	github.com/denverdino/aliyungo v0.0.0-20200720072455-26fa39a46424
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/client_golang v0.9.3
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
)

replace (
  github.com/allanhung/alicloud-monitoring/pkg/alicloud v0.1.0 => pkg/alicloud v0.1.0
)
