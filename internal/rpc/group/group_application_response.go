package group

import (
	"Open_IM/internal/rpc/chat"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/proto/group"
	"context"
)

func (s *groupServer) GroupApplicationResponse(_ context.Context, pb *group.GroupApplicationResponseReq) (*group.GroupApplicationResponseResp, error) {
	log.NewInfo(pb.OperationID, "GroupApplicationResponse args: ", pb.String())
	reply, err := im_mysql_model.GroupApplicationResponse(pb)
	if err != nil {
		log.NewError(pb.OperationID, "GroupApplicationResponse failed ", err.Error(), pb)
		return &group.GroupApplicationResponseResp{ErrCode: 702, ErrMsg: err.Error()}, nil
	}

	if pb.HandleResult == 1 {
		if pb.ToUserID == "0" {
			err = db.DB.AddGroupMember(pb.GroupID, pb.FromUserID)
			if err != nil {
				log.Error("", "", "rpc GroupApplicationResponse call..., db.DB.AddGroupMember fail [pb: %s] [err: %s]", pb.String(), err.Error())
				return nil, err
			}
		} else {
			err = db.DB.AddGroupMember(pb.GroupID, pb.ToUserID)
			if err != nil {
				log.Error("", "", "rpc GroupApplicationResponse call..., db.DB.AddGroupMember fail [pb: %s] [err: %s]", pb.String(), err.Error())
				return nil, err
			}
		}
	}
	if pb.ToUserID == "0" {
		chat.ApplicationProcessedNotification(pb.OperationID, pb.FromUserID)
	}

	if pb.HandleResult == 1 {

	}

	log.NewInfo(pb.OperationID, "rpc GroupApplicationResponse ok ", reply)

	return reply, nil
}
