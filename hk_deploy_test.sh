#!/bin/bash

set -e

pull_dir=/home/go/src
push_dir=/home/apps

godemo=godemo

date=`date +%F_%T`

cd `dirname $0`;pwd
help(){
echo "1.测试服ccpay ./deploy_ec_test.sh application"
}

godemo(){
echo -n "$date"
echo "building godemo"
cd $pull_dir/$godemo
go build

md5sum $pull_dir/$godemo/$godemo
rm -rf $push_dir/$godemo/$godemo
echo "copy"
cp -r $pull_dir/$godemo/$godemo $push_dir/$godemo

echo "copy templates"
cp -R $pull_dir/$godemo/templates $push_dir/$godemo 

supervisorctl restart godemo
}

case $1 in
godemo ) godemo;;
help ) help ;;
* ) echo "USAGE: godemo" ;;
esac