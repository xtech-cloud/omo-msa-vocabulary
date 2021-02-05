export GO111MODULE=on
export GOSUMDB=off
export GOPROXY=https://mirrors.aliyun.com/goproxy/
go install omo.msa.vocabulary
mkdir _build
mkdir _build/bin

cp -rf /root/go/bin/omo.msa.vocabulary _build/bin/
cp -rf conf _build/
cd _build
tar -zcf msa.vocabulary.tar.gz ./*
mv msa.vocabulary.tar.gz ../
cd ../
rm -rf _build
