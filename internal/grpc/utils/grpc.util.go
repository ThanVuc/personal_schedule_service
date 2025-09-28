package utils

import (
	"math"
	"personal_schedule_service/internal/grpc/models"
	"personal_schedule_service/proto/common"
)

func ToPagination(pageQuery *common.PageQuery) models.Pagination {
	if pageQuery == nil {
		return models.Pagination{
			Limit:  0,
			Offset: 0,
		}
	}

	if pageQuery.PageIgnore == nil || *pageQuery.PageIgnore {
		return models.Pagination{
			Limit:  0,
			Offset: 0,
		}
	}

	if pageQuery.PageSize <= 0 {
		pageQuery.PageSize = 10
	}
	if pageQuery.Page <= 0 {
		pageQuery.Page = 1
	}

	return models.Pagination{
		Limit:  pageQuery.PageSize,
		Offset: (pageQuery.Page - 1) * pageQuery.PageSize,
	}
}

func ToPageInfo(page, pageSize, totalItems int32) *common.PageInfo {
	totalPages := int32(math.Ceil(float64(totalItems) / float64(pageSize)))

	return &common.PageInfo{
		TotalItems: totalItems,
		Page:       page,
		TotalPages: totalPages,
		PageSize:   pageSize,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	}
}
