pipeline {
    agent any
    stages {
        stage('build') {
            agent {
                dockerfile { 
                    dir 'ci' 
                    args '-v $WORKSPACE:/go/src/git.townsourced.com/townsourced/townsourced'
                }
            }
            environment {
                GOPATH = '/go'
                HOME = '.'
            }
            steps {
                sh '''
                    cd $GOPATH/src/git.townsourced.com/townsourced/townsourced
                    go clean -i -a
                    go build -o townsourced
                    cd web 

                    npm install
                    gobble build static -f

                    cd ../..
                    rm -rf release
                    mkdir -p release/usr/local/bin

                    mv townsourced release/usr/local/bin/
                    mv web/static release/usr/local/share/townsourced/web/

                    tar -czf release.tar.gz release/
                '''

                archiveArtifacts artifacts: 'release.tar.gz'
            }
        }
        stage('deploy') {
            steps {
                withCredentials([sshUserPrivateKey(credentialsId: 'redacted', keyFileVariable: 'KEY_FILE')]) {
                    sh '''
                        scp -i $KEY_FILE release.tar.gz tshannon@redacted:/home/tshannon/workspace/releases/ 
                        ssh -i $KEY_FILE -tt tshannon@redacted "
                            set -e
                            cd /home/tshannon/workspace/releases/
                            tar -xzf release.tar.gz
                            sudo chown -R root:root release
                            sudo systemctl stop townsourced.service
                            sudo mv /usr/local/share/townsourced/web/bin release/usr/local/share/townsourced/web/bin
                            sudo rm -rf /usr/local/share/townsourced/
                            sudo mv release/usr/local/share/townsourced /usr/local/share/townsourced
                            sudo rm -rf /usr/local/bin/townsourced
                            sudo mv release/usr/local/bin/townsourced /usr/local/bin/townsourced
                            sudo systemctl start townsourced.service
                            sudo rm -rf release
                            sudo rm release.tar.gz
                        "
                    '''
                }
            }
        }
    }
}

