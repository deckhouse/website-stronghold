{{- define "resources" }}
resources:
  requests:
    memory: {{ pluck .Values.werf.env .Values.resources.requests.memory | first | default .Values.resources.requests.memory._default }}
{{- end }}

{{- define "readiness_probe" }}
failureThreshold: 5
periodSeconds: 10
timeoutSeconds: 5
{{- end }}
{{- define "liveness_probe" }}
failureThreshold: 10
periodSeconds: 10
timeoutSeconds: 5
{{- end }}
{{- define "startup_probe" }}
failureThreshold: 10
periodSeconds: 10
timeoutSeconds: 5
{{- end }}
