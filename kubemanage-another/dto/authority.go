package dto

import "github.com/noovertime7/kubemanage/dao/model"

type AuthorityList struct {
	PageInfo
	Total             int64                `json:"total"`
	AuthorityListItem []model.SysAuthority `json:"list"`
}
type AuthorityInput struct {
	AuthorityId   uint   `json:"authorityId"`
	AuthorityName string `json:"authorityName"`
	ParentId      uint   `json:"parentId"`
	DefaultRouter string `json:"defaultRouter"`
}
