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
    stages {
        stage('make geth') {
            steps {
                script {
                    sh 'make geth'
                }
            }
        }
    }
    post {
        always{
            script {
                currentBuild.displayName = "${env.BUILD_NUMBER}: ${APP_NAME}:${VERSION}"
            }
        }
        success {
          archiveArtifacts artifacts: 'build/bin/geth', followSymlinks: false
        }
    }
}
