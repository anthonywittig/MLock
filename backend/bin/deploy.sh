#!/bin/bash

set -ex

handleLambda () {
    lambda=$1
    buildDir="build/${lambda}"

    mkdir -p $buildDir # some might throw a few extra files in here before they call this function
    if test -f "${lambda}/.env"; then
        cp ${lambda}/.env $buildDir
    fi
    GOOS=linux GOARCH=amd64 go build -o $buildDir ${lambda}/main.go
    (cd $buildDir && zip -r function.zip .* *) # .* for hidden files
    aws --profile=mLock lambda update-function-code \
        --function-name $lambda \
        --zip-file fileb://${buildDir}/function.zip
}

handleDb () {
    cd $scriptDir
    cd ../db

    rm -rf build
    mkdir build

    mkdir -p build/migrations/goosemigrations
    cp -r migrations/goosemigrations/ build/migrations/goosemigrations

    handleLambda migrations
    # TODO: execute lambda with "forward" option to run the migrations forward
}

handleApis () {
    cd $scriptDir
    cd ../apis

    rm -rf build
    mkdir build

    #lambdas=('helloworld' 'helloworld2' 'users')
    lambdas=('users')
    for lambda in "${lambdas[@]}" ; do
        handleLambda $lambda
    done
}

scriptDir=$(dirname "$0")
cd $scriptDir
scriptDir=$(pwd)

handleDb

handleApis

