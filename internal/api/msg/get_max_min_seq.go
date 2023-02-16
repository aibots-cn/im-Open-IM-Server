package msg

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/tokenverify"
	pbChat "Open_IM/pkg/proto/msg"
	sdkws "Open_IM/pkg/proto/sdkws"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type paramsUserNewestSeq struct {
	ReqIdentifier int    `json:"reqIdentifier" binding:"required"`
	SendID        string `json:"sendID" binding:"required"`
	OperationID   string `json:"operationID" binding:"required"`
	MsgIncr       int    `json:"msgIncr" binding:"required"`
}

func GetSeq(c *gin.Context) {
	params := paramsUserNewestSeq{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	token := c.Request.Header.Get("token")
	if ok, err := tokenverify.VerifyToken(token, params.SendID); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token validate err" + err.Error()})
		return
	}
	pbData := sdkws.GetMaxAndMinSeqReq{}
	pbData.UserID = params.SendID
	pbData.OperationID = params.OperationID
	grpcConn := rpc.GetDefaultConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMsgName, pbData.OperationID)
	if grpcConn == nil {
		errMsg := pbData.OperationID + " getcdv3.GetDefaultConn == nil"
		log.NewError(pbData.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}

	msgClient := pbChat.NewMsgClient(grpcConn)
	reply, err := msgClient.GetMaxAndMinSeq(context.Background(), &pbData)
	if err != nil {
		log.NewError(params.OperationID, "UserGetSeq rpc failed, ", params, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "UserGetSeq rpc failed, " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errCode":       reply.ErrCode,
		"errMsg":        reply.ErrMsg,
		"msgIncr":       params.MsgIncr,
		"reqIdentifier": params.ReqIdentifier,
		"data": gin.H{
			"maxSeq": reply.MaxSeq,
			"minSeq": reply.MinSeq,
		},
	})

}
