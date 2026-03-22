pipeline {
  agent any

  environment {
    MODULE = 'github.com/aptlogica/sereni-storage-provider'
    COVER_PROFILE = 'coverage.out'
    COVER_HTML = 'coverage.html'
  }

  stages {
    stage('Checkout Code') {
      steps {
        checkout scm
      }
    }

    stage('Test & Coverage') {
      steps {
        sh 'go test -v -coverprofile=$COVER_PROFILE -covermode=atomic -coverpkg=$MODULE/... $MODULE/tests/...'
        sh 'go tool cover -html=$COVER_PROFILE -o $COVER_HTML'
      }
    }

    stage('SonarQube Analysis') {
        when {
        anyOf {
          branch 'develop'
          branch 'main'
          branch 'release/*'
          branch 'master'
        }
      }
      steps {
        script {
          // Get path to the installed Sonar Scanner tool
          def scannerHome = tool 'SonarScanner'

          withSonarQubeEnv('aptl-sonar') {
            // Run the scanner binary
            sh "${scannerHome}/bin/sonar-scanner"
          }
        }
      }
    }

    stage('Quality Gate') {
      when {
        anyOf {
          branch 'develop'
          branch 'main'
          branch 'release/*'
          branch 'master'
        }
      }
      steps {
        timeout(time: 10, unit: 'MINUTES') {
          waitForQualityGate abortPipeline: true
        }
      }
    }
  }
}
