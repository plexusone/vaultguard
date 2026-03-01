# helm-agent-base

A reusable Helm chart for deploying ADK/Eino-based AI agents with OmniSafe security integration.

## Why a Generic Chart?

The `stats-agent-team` Helm chart has agent-specific logic that could be generalized:

| Current (stats-agent-team) | Generic (helm-agent-base) |
|---------------------------|---------------------------|
| Hardcoded 5 agents | Configurable agent list |
| Stats-specific config | Generic agent config |
| Mixed concerns | Separation of concerns |

## Architecture

```
helm-agent-base/                    # Generic chart (new repo)
├── Chart.yaml
├── values.yaml                     # Base defaults
├── values-security.yaml            # OmniSafe defaults
├── templates/
│   ├── _helpers.tpl
│   ├── _agent.tpl                  # Agent deployment template
│   ├── configmap.yaml              # Shared config
│   ├── secret.yaml                 # Credentials (dev)
│   ├── serviceaccount.yaml         # With IRSA support
│   ├── security-configmap.yaml     # OmniSafe config
│   └── agents/                     # Per-agent resources
│       └── {{range .Values.agents}}
│           ├── deployment.yaml
│           ├── service.yaml
│           └── hpa.yaml
│       {{end}}
└── charts/                         # Subcharts
    └── omnisafe/                   # OmniSafe configuration subchart

stats-agent-team/                   # Uses helm-agent-base as dependency
├── Chart.yaml                      # depends on helm-agent-base
├── values.yaml                     # Stats-specific values
└── templates/                      # Stats-specific overrides only
```

## Usage

### As a Dependency (Recommended)

```yaml
# stats-agent-team/Chart.yaml
apiVersion: v2
name: stats-agent-team
version: 1.0.0
dependencies:
  - name: agent-base
    version: "1.0.0"
    repository: "https://charts.plexusone.io"
```

```yaml
# stats-agent-team/values.yaml
agent-base:
  # Agent definitions
  agents:
    research:
      enabled: true
      image: ghcr.io/plexusone/stats-agent-team:latest
      command: ["/app/research"]
      ports:
        http: 8001
        a2a: 9001

    synthesis:
      enabled: true
      image: ghcr.io/plexusone/stats-agent-team:latest
      command: ["/app/synthesis"]
      ports:
        http: 8004

    verification:
      enabled: true
      image: ghcr.io/plexusone/stats-agent-team:latest
      command: ["/app/verification"]
      ports:
        http: 8002

    orchestration:
      enabled: true
      image: ghcr.io/plexusone/stats-agent-team:latest
      command: ["/app/orchestration-eino"]
      ports:
        http: 8000
        a2a: 9000

  # LLM Configuration
  llm:
    provider: gemini

  # Search Configuration
  search:
    provider: serper

  # OmniSafe Security
  security:
    enabled: true
    local:
      minSecurityScore: 50
    cloud:
      aws:
        requireIRSA: true
```

### Standalone

```yaml
# my-agents/values.yaml
agents:
  chatbot:
    enabled: true
    image: myregistry/chatbot:v1
    command: ["/app/chatbot"]
    ports:
      http: 8080
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 256Mi

  summarizer:
    enabled: true
    image: myregistry/summarizer:v1
    command: ["/app/summarizer"]
    ports:
      http: 8081

llm:
  provider: claude

security:
  enabled: true
  cloud:
    requireIAM: true
```

## Generic Agent Template

```yaml
# templates/_agent.tpl
{{- define "agent-base.agent" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "agent-base.fullname" . }}-{{ .agentName }}
spec:
  replicas: {{ .agent.replicas | default 1 }}
  template:
    spec:
      serviceAccountName: {{ include "agent-base.serviceAccountName" . }}
      containers:
        - name: {{ .agentName }}
          image: {{ .agent.image }}
          command: {{ .agent.command | toYaml | nindent 12 }}
          ports:
            {{- range $name, $port := .agent.ports }}
            - name: {{ $name }}
              containerPort: {{ $port }}
            {{- end }}
          envFrom:
            - configMapRef:
                name: {{ include "agent-base.fullname" . }}-config
            - configMapRef:
                name: {{ include "agent-base.fullname" . }}-security
            {{- if .Values.secrets.create }}
            - secretRef:
                name: {{ include "agent-base.fullname" . }}-secrets
            {{- end }}
          resources:
            {{- toYaml .agent.resources | nindent 12 }}
{{- end -}}
```

## OmniSafe Security ConfigMap

```yaml
# templates/security-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "agent-base.fullname" . }}-security
data:
  # OmniSafe configuration
  OMNISAFE_ENABLED: {{ .Values.security.enabled | quote }}

  # Local security
  OMNISAFE_MIN_SECURITY_SCORE: {{ .Values.security.local.minSecurityScore | quote }}
  OMNISAFE_REQUIRE_ENCRYPTION: {{ .Values.security.local.requireEncryption | quote }}
  OMNISAFE_REQUIRE_TPM: {{ .Values.security.local.requireTPM | quote }}

  # Cloud security
  OMNISAFE_REQUIRE_IAM: {{ .Values.security.cloud.requireIAM | quote }}
  OMNISAFE_REQUIRE_IRSA: {{ .Values.security.cloud.aws.requireIRSA | quote }}
  OMNISAFE_ALLOWED_ROLE_ARNS: {{ .Values.security.cloud.aws.allowedRoleARNs | join "," | quote }}

  # Kubernetes security
  OMNISAFE_DENIED_NAMESPACES: {{ .Values.security.kubernetes.deniedNamespaces | join "," | quote }}
  OMNISAFE_ALLOWED_NAMESPACES: {{ .Values.security.kubernetes.allowedNamespaces | join "," | quote }}

  # Development
  OMNISAFE_ALLOW_INSECURE: {{ .Values.security.development.allowInsecure | quote }}
```

## Benefits

1. **Reusability** - One chart for all agent projects
2. **Consistency** - Same security, observability, and deployment patterns
3. **Maintainability** - Update base chart, all projects benefit
4. **Flexibility** - Override anything via values
5. **Security by Default** - OmniSafe integrated from the start

## Migration Path for stats-agent-team

1. Create `helm-agent-base` as new repo
2. Extract generic templates from stats-agent-team
3. Add helm-agent-base as dependency
4. Move stats-specific config to values.yaml
5. Test with existing deployment

## Proposed Repository Structure

```
github.com/plexusone/
├── omnisafe/           # Security integration library
├── omnitrust/          # Security posture assessment
├── omnivault/          # Secret management
├── helm-agent-base/    # Generic agent Helm chart (NEW)
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
└── stats-agent-team/   # Uses helm-agent-base
    ├── Chart.yaml      # depends on helm-agent-base
    └── values.yaml     # Stats-specific config
```
