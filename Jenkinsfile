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
      docker { image 'golang:1.19' }
    }
    stages {
        stage('prep env') {
          steps {
            script {
              sh 'apk add make'
            }
          }
        }
        stage('make geth') {
            steps {
                script {
                    sh 'make geth'
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
