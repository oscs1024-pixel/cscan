package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cscan/api/internal/logic/common"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type AssetInventoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssetInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssetInventoryLogic {
	return &AssetInventoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// buildInventoryFilter 构建资产清单的查询条件
func (l *AssetInventoryLogic) buildInventoryFilter(req *types.AssetInventoryReq) bson.M {
	filter := bson.M{}

	// 搜索关键词
	if req.Query != "" {
		q := req.Query
		filter["$or"] = []bson.M{
			{"host": bson.M{"$regex": q, "$options": "i"}},
			{"title": bson.M{"$regex": q, "$options": "i"}},
			{"domain": bson.M{"$regex": q, "$options": "i"}},
			{"ip.ipv4.ip": bson.M{"$regex": q, "$options": "i"}},
			{"ip.ipv6.ip": bson.M{"$regex": q, "$options": "i"}},
		}
	}

	// 域名过滤
	if req.Domain != "" {
		filter["host"] = bson.M{"$regex": req.Domain, "$options": "i"}
	}

	// 端口过滤
	if len(req.Ports) > 0 {
		filter["port"] = bson.M{"$in": req.Ports}
	}

	// 状态码过滤
	if len(req.StatusCodes) > 0 {
		filter["status"] = bson.M{"$in": req.StatusCodes}
	}

	// 标签过滤
	if len(req.Labels) > 0 {
		filter["labels"] = bson.M{"$in": req.Labels}
	}

	// 服务类型过滤
	if req.Service != "" {
		filter["service"] = bson.M{"$regex": req.Service, "$options": "i"}
	}

	// IconHash 过滤
	if req.IconHash != "" {
		filter["icon_hash"] = req.IconHash
	}

	// 技术栈过滤
	if len(req.Technologies) > 0 {
		techFilters := make([]bson.M, 0, len(req.Technologies))
		for _, tech := range req.Technologies {
			escapedTech := regexp.QuoteMeta(tech)
			techFilters = append(techFilters, bson.M{
				"app": bson.M{"$regex": escapedTech, "$options": "i"},
			})
		}
		if len(techFilters) > 0 {
			if existingOr, ok := filter["$or"]; ok {
				filter["$and"] = []bson.M{
					{"$or": existingOr},
					{"$or": techFilters},
				}
				delete(filter, "$or")
			} else {
				filter["$or"] = techFilters
			}
		}
	}

	// 时间范围过滤
	if req.TimeRange != "" && req.TimeRange != "all" {
		now := time.Now()
		var startTime time.Time
		switch req.TimeRange {
		case "24h":
			startTime = now.Add(-24 * time.Hour)
		case "7d":
			startTime = now.Add(-7 * 24 * time.Hour)
		case "30d":
			startTime = now.Add(-30 * 24 * time.Hour)
		}
		if !startTime.IsZero() {
			filter["update_time"] = bson.M{"$gte": startTime}
		}
	}

	return filter
}

// convertAssetToInventoryItem 将 Asset 模型转换为清单展示项
func convertAssetToInventoryItem(asset model.Asset, wsId string) types.AssetInventoryItem {
	ip := ""
	var ips []string
	if len(asset.Ip.IpV4) > 0 {
		ip = asset.Ip.IpV4[0].IPName
	} else if len(asset.Ip.IpV6) > 0 {
		ip = asset.Ip.IpV6[0].IPName
	}
	for _, v4 := range asset.Ip.IpV4 {
		ips = append(ips, v4.IPName)
	}
	for _, v6 := range asset.Ip.IpV6 {
		ips = append(ips, v6.IPName)
	}

	iconHashBytes := ""
	if len(asset.IconHashBytes) > 0 && isValidImageBytes(asset.IconHashBytes) {
		iconHashBytes = base64.StdEncoding.EncodeToString(asset.IconHashBytes)
	}

	labels := asset.Labels
	if labels == nil {
		labels = []string{}
	}

	return types.AssetInventoryItem{
		Id:              asset.Id.Hex(),
		WorkspaceId:     wsId,
		Host:            asset.Host,
		IP:              ip,
		Ips:             ips,
		Port:            asset.Port,
		Service:         asset.Service,
		Title:           asset.Title,
		Technologies:    asset.App,
		Labels:          labels,
		Status:          asset.HttpStatus,
		Domain:          asset.Domain,
		LastUpdated:     formatTimeAgo(asset.UpdateTime),
		FirstSeen:       asset.CreateTime.Local().Format("2006-01-02 15:04:05"),
		LastUpdatedFull: asset.UpdateTime.Local().Format("2006-01-02 15:04:05"),
		Screenshot:      asset.Screenshot,
		IconHash:        asset.IconHash,
		IconHashBytes:   iconHashBytes,
		HttpHeader:      asset.HttpHeader,
		HttpBody:        asset.HttpBody,
		Banner:          asset.Banner,
		CName:           asset.CName,
	}
}

// AssetInventory 获取资产清单
func (l *AssetInventoryLogic) AssetInventory(req *types.AssetInventoryReq, workspaceId string) (resp *types.AssetInventoryResp, err error) {
	l.Logger.Infof("AssetInventory查询: workspaceId=%s, page=%d, pageSize=%d", workspaceId, req.Page, req.PageSize)

	// 获取需要查询的工作空间列表
	wsIds := common.GetWorkspaceIds(l.ctx, l.svcCtx, workspaceId)
	l.Logger.Infof("AssetInventory查询工作空间列表: %v", wsIds)

	// 构建查询条件
	filter := l.buildInventoryFilter(req)

	var total int64
	var resultItems []types.AssetInventoryItem

	if len(wsIds) > 1 || workspaceId == "" || workspaceId == "all" {
		// 多工作空间：需要获取足够数据以覆盖当前页
		needTotal := req.Page * req.PageSize // 需要覆盖到当前页末尾的数据量
		var allItems []types.AssetInventoryItem

		for _, wsId := range wsIds {
			assetModel := l.svcCtx.GetAssetModel(wsId)

			wsTotal, err := assetModel.Count(l.ctx, filter)
			if err != nil || wsTotal == 0 {
				continue
			}
			total += wsTotal

			limit := needTotal
			if wsTotal < int64(limit) {
				limit = int(wsTotal)
			}

			assets, err := assetModel.FindWithScreenshot(l.ctx, filter, 1, limit)
			if err != nil {
				l.Logger.Errorf("查询工作空间 %s 资产失败: %v", wsId, err)
				continue
			}

			for _, asset := range assets {
				allItems = append(allItems, convertAssetToInventoryItem(asset, wsId))
			}
		}

		// 排序
		sortAssets(allItems, req.SortBy)

		// 分页
		start := (req.Page - 1) * req.PageSize
		end := start + req.PageSize
		if start >= len(allItems) {
			resultItems = []types.AssetInventoryItem{}
		} else {
			if end > len(allItems) {
				end = len(allItems)
			}
			resultItems = allItems[start:end]
		}
	} else {
		// 单工作空间：使用 MongoDB 原生分页，避免全量加载
		wsId := wsIds[0]
		assetModel := l.svcCtx.GetAssetModel(wsId)

		wsTotal, err := assetModel.Count(l.ctx, filter)
		if err != nil {
			return &types.AssetInventoryResp{Code: 500, Msg: "查询失败"}, nil
		}
		total = wsTotal

		sortField := inventorySortField(req.SortBy)
		skip := int64((req.Page - 1) * req.PageSize)
		limit := int64(req.PageSize)
		assets, err := assetModel.FindWithOffset(l.ctx, filter, skip, limit, sortField)
		if err != nil {
			return &types.AssetInventoryResp{Code: 500, Msg: "查询失败"}, nil
		}

		resultItems = make([]types.AssetInventoryItem, 0, len(assets))
		for _, asset := range assets {
			resultItems = append(resultItems, convertAssetToInventoryItem(asset, wsId))
		}
	}

	return &types.AssetInventoryResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  resultItems,
	}, nil
}

// inventorySortField 将前端排序参数转为 MongoDB 排序字段
// 返回值带 "-" 前缀表示降序，无前缀表示升序
func inventorySortField(sortBy string) string {
	switch sortBy {
	case "name", "name-asc":
		return "host"
	case "name-desc":
		return "-host"
	case "time-asc":
		return "update_time"
	case "port":
		return "port"
	default: // "time", "time-desc", ""
		return "-update_time"
	}
}

// sortAssets 对资产进行排序
func sortAssets(assets []types.AssetInventoryItem, sortBy string) {
	switch sortBy {
	case "name", "name-asc":
		// 按主机名升序
		for i := 0; i < len(assets)-1; i++ {
			for j := i + 1; j < len(assets); j++ {
				if strings.ToLower(assets[i].Host) > strings.ToLower(assets[j].Host) {
					assets[i], assets[j] = assets[j], assets[i]
				}
			}
		}
	case "name-desc":
		// 按主机名降序
		for i := 0; i < len(assets)-1; i++ {
			for j := i + 1; j < len(assets); j++ {
				if strings.ToLower(assets[i].Host) < strings.ToLower(assets[j].Host) {
					assets[i], assets[j] = assets[j], assets[i]
				}
			}
		}
	case "port":
		// 按端口升序
		for i := 0; i < len(assets)-1; i++ {
			for j := i + 1; j < len(assets); j++ {
				if assets[i].Port > assets[j].Port {
					assets[i], assets[j] = assets[j], assets[i]
				}
			}
		}
	case "time-asc":
		// 按时间升序（最旧的在前）
		for i := 0; i < len(assets)-1; i++ {
			for j := i + 1; j < len(assets); j++ {
				if assets[i].LastUpdatedFull > assets[j].LastUpdatedFull {
					assets[i], assets[j] = assets[j], assets[i]
				}
			}
		}
	}
	// 默认 "time"/"time-desc"/空值：按时间降序（已经是最新的在前）
}

// formatTimeAgo 格式化相对时间
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	} else {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1天前"
		}
		return fmt.Sprintf("%d天前", days)
	}
}
