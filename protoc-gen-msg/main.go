package main

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/davyxu/golog"
	"github.com/davyxu/pbmeta"
	pbprotos "github.com/davyxu/pbmeta/proto"
	plugin "github.com/davyxu/pbmeta/proto/compiler"
	"github.com/golang/protobuf/proto"
)

var log *golog.Logger = golog.New("main")

func main() {
	//读取protoc请求
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Errorln("reading input")
	}

	var Request plugin.CodeGeneratorRequest   //The input.
	var Response plugin.CodeGeneratorResponse //The output.

	//解析请求
	if err := proto.Unmarshal(data, &Request); err != nil {
		log.Errorln("parsing input proto")
	}

	if len(Request.FileToGenerate) == 0 {
		log.Errorln("no files to generate")
	}

	if len(Request.ProtoFile) == 0 {
		log.Errorln("no proto file")
	}

	var packageName string
	for i, file := range Request.ProtoFile {
		if i == 0 {
			packageName = *file.Package
		}else {
			if packageName != *file.Package {
				log.Errorln("could handle ONLY ONE PACKAGE for all protos")
			}
		}
	}

	//建立解析池
	pool := pbmeta.NewDescriptorPool(&pbprotos.FileDescriptorSet{
		File: Request.ProtoFile,
	})

	Response.File = make([]*plugin.CodeGeneratorResponse_File, 0)

	context, ok := printFile(pool, packageName)
	if !ok {
		os.Exit(1)
	}

	Response.File = append(Response.File, &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(Request.GetParameter()),
		Content: proto.String(context),
	})

	//发回处理结果
	data, err = proto.Marshal(&Response)
	if err != nil {
		log.Errorln("failed to marshal output proto")
	}

	//写入文件
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Errorln("failed to write output proto")
	}
}

func changeExt(name string) string {
	ext := path.Ext(name)
	if ext == ".proto" {
		name = name[0 : len(name)-len(ext)]
	}
	return name + ".msg.go"
}
