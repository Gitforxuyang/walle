package server

import (
	"context"
	"fmt"
	"github.com/Gitforxuyang/protoreflect/metadata"
	"github.com/Gitforxuyang/walle/client/grpc"
	error2 "github.com/Gitforxuyang/walle/util/error"
	"github.com/Gitforxuyang/walle/util/utils"
)

//单点登录
func checkSSO(ctx context.Context, token string) (uid string, err error) {
	proxy := grpc.GetProxy("user")
	svc := grpc.GetService("user")

	md := make(metadata.Metadata)
	resp, err := proxy.Invoke(ctx, fmt.Sprintf("%s.%s", svc.Package, svc.Name), "checkSSO",
		[]byte(fmt.Sprintf(`{"token":"%s"}`, token)), &md)
	if err != nil {
		return "", err
	}
	r, err := utils.JsonToMap(string(resp))
	if err != nil {
		return "", error2.UnknowError.SetDetail(err.Error())
	}
	uid = r["uid"].(string)
	return uid, nil
}
