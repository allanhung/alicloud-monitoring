{{- define "alicloudmonitoring.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "alicloudmonitoring.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "alicloudmonitoring.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "alicloudmonitoring.matchLabels" -}}
app: {{ template "alicloudmonitoring.name" . }}
release: {{ .Release.Name }}
{{- end -}}

{{- define "alicloudmonitoring.metaLabels" -}}
version: {{ .Chart.AppVersion }}
chart: {{ template "alicloudmonitoring.chart" . }}
{{- end -}}

{{- define "alicloudmonitoring.labels" -}}
{{ include "alicloudmonitoring.matchLabels" . }}
{{ include "alicloudmonitoring.metaLabels" . }}
{{- end -}}

{{- define "podMonitor.namespaceSelector" -}}
{{- if .Values.podMonitor.namespaceSelector -}}
{{ toYaml .Values.podMonitor.namespaceSelector -}}
{{- else -}}
{{ (.Values.namespace | list) -}}
{{- end -}}
{{- end -}}

{{- define "podMonitor.labelSelector" -}}
{{- if .Values.podMonitor.labelSelector -}}
{{ toYaml .Values.podMonitor.labelSelector -}}
{{- else -}}
{{ include "alicloudmonitoring.matchLabels" . }}
{{- end -}}
{{- end -}}
