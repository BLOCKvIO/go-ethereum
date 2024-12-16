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
    triggers {
      cron 'H 0 29 * *'
    }
    environment {
      GOCACHE = '/tmp/'
      HOME = "$WORKSPACE"
    }
    options {
        disableConcurrentBuilds()
        ansiColor('xterm')
    }
    agent {
      docker { image 'golang:1.19' }
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
        sh 'echo ok'
          archiveArtifacts artifacts: 'build/bin/geth', followSymlinks: false
        slackSend color: "00FF00", teamDomain: 'smartmedialabs', tokenCredentialId: 'vl-slack-token' , channel: '#web3_builds', message: "geth(${env.BRANCH_NAME}) Built successfully (<${env.BUILD_URL}|Open>)"
      }
      unstable {
        sh 'echo ok'
        slackSend color: "00FF00", teamDomain: 'smartmedialabs', tokenCredentialId: 'vl-slack-token', channel: '#web3_builds', message: "geth(${env.BRANCH_NAME}) Built successfully (<${env.BUILD_URL}|Open>)"
      }
      failure {
        sh 'echo error'
        slackSend color: "FF0000", teamDomain: 'smartmedialabs', tokenCredentialId: 'vl-slack-token', channel: '#web3_builds', message: "geth(${env.BRANCH_NAME}) - Build failure [${ERROR}] (<${env.BUILD_URL}|Open>)"
      }
    }
}
