{{ "{{- if .Values.metrics.service.create }}" }}
apiVersion: v1
kind: Service
metadata:
  name: {{ "{{ .Chart.Name | trimSuffix \"-chart\" | trunc 44 }}-controller-metrics" }}
  namespace: {{ "{{ .Release.Namespace }}" }}
  labels:
    app.kubernetes.io/name: {{ IncludeTemplate "app.name" }}
    app.kubernetes.io/instance: {{ "{{ .Release.Name }}" }}
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/version: {{ "{{ .Chart.AppVersion | quote }}" }}
    k8s-app: {{ IncludeTemplate "app.name" }}
    helm.sh/chart: {{ IncludeTemplate "chart.name-version" }}
spec:
  selector:
    app.kubernetes.io/name: {{ IncludeTemplate "app.name" }}
    app.kubernetes.io/instance: {{ "{{ .Release.Name }}" }}
    app.kubernetes.io/managed-by: Helm
    k8s-app: {{ IncludeTemplate "app.name" }}
{{ "{{- range $key, $value := .Values.deployment.labels }}" }}
    {{ "{{ $key }}: {{ $value | quote }}" }}
{{ "{{- end }}" }}
  type: {{ "{{ .Values.metrics.service.type }}" }}
  ports:
  - name: metricsport
    port: 8080
    targetPort: http
    protocol: TCP
{{ "{{- end }}" }}
