package models

import (
	"git/inspursoft/board/src/adminserver/models/nodeModel"
	"github.com/astaxie/beego/orm"
)

func RegisterModels() {
	orm.RegisterModel(
		new(nodeModel.NodeLog),
		new(nodeModel.NodeLogDetailInfo),
		new(nodeModel.NodeStatus))
}