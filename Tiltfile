version_settings(constraint='>=0.33.0')
load('ext://restart_process', 'docker_build_with_restart')

# Namespace and secrets
k8s_yaml('k8s/namespace.yaml')
k8s_yaml('k8s/secret.yaml')

# Infrastructure
k8s_yaml([
    'k8s/postgres/init-configmap.yaml',
    'k8s/postgres/deployment.yaml',
    'k8s/postgres/service.yaml',
])
k8s_yaml([
    'k8s/redis/deployment.yaml',
    'k8s/redis/service.yaml',
])

# Services
# Live update strategy: local_resource builds the Linux binary on the host (workspace-aware),
# then live_update syncs the compiled binary into the running alpine container and restarts.
services = [
    ('api-gateway',      'services/api-gateway',      8080),
    ('user-service',     'services/user-service',      8081),
    ('workout-service',  'services/workout-service',   8082),
    ('exercise-service', 'services/exercise-service',  8083),
]

for name, path, port in services:
    # Include all of services/ so every Dockerfile has the go.mod/go.sum files it needs for `go work sync`
    watch_paths = ['go.work', 'go.work.sum', 'services/', 'shared/']

    local_resource(
        name + '-build',
        cmd = 'go build -o bin/' + name + ' ./' + path + '/cmd/...',
        env = {
            'CGO_ENABLED': '0',
            'GOOS': 'linux',
            'GOARCH': 'amd64',
        },
        deps = watch_paths,
        labels = ['build'],
    )

    docker_build_with_restart(
        name,
        context = '.',
        dockerfile = path + '/Dockerfile',
        only = watch_paths,
        entrypoint = '/app',
        live_update = [
            sync('./bin/' + name, '/app'),
        ],
    )

    k8s_yaml([
        'k8s/' + name + '/deployment.yaml',
        'k8s/' + name + '/service.yaml',
    ])

# Port forwards and dependency ordering
k8s_resource('postgres', port_forwards = '5432:5432', labels = ['infra'])
k8s_resource('redis',    port_forwards = '6379:6379', labels = ['infra'])

k8s_resource(
    'user-service',
    port_forwards = '8081:8081',
    resource_deps = ['postgres'],
    labels = ['services'],
)
k8s_resource(
    'exercise-service',
    port_forwards = '8083:8083',
    resource_deps = ['postgres', 'redis'],
    labels = ['services'],
)
k8s_resource(
    'workout-service',
    port_forwards = '8082:8082',
    resource_deps = ['postgres', 'redis'],
    labels = ['services'],
    # workout-service cmd/main.go is an empty stub — pod exits immediately
    trigger_mode = TRIGGER_MODE_MANUAL,
)
k8s_resource(
    'api-gateway',
    port_forwards = '8080:8080',
    resource_deps = ['user-service', 'exercise-service'],
    labels = ['services'],
)
