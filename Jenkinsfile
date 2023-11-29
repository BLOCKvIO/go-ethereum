node {
    properties(
        [
            [
                $class: 'GithubProjectProperty',
                displayName: 'k8s-config',
                projectUrlStr: 'https://github.com/BLOCKvIO/go-ethereum/'
            ]
        ]
    )
}
pipeline {
    options {
        disableConcurrentBuilds()
        ansiColor('xterm')
    }
    agent {
      docker { image 'golang:1.19-alpine' }
    }
    stages {
        stage('build geth') {
            steps {
                script {
                    sh 'env GO111MODULE=on go run build/ci.go install ./cmd/geth'
                }
            }
        }
    }
    post {
        success {
          archiveArtifacts artifacts: 'build/bin/geth', followSymlinks: false
        }
    }
}
