package controller

import (
	"Open_IM/pkg/common/db/cache"
	"Open_IM/pkg/common/db/relation"
	"Open_IM/pkg/common/db/unrelation"
	"context"
	"github.com/dtm-labs/rockscache"
	_ "github.com/dtm-labs/rockscache"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type GroupInterface interface {
	FindGroupsByID(ctx context.Context, groupIDs []string) (groups []*relation.Group, err error)
	CreateGroup(ctx context.Context, groups []*relation.Group) error
	DeleteGroupByIDs(ctx context.Context, groupIDs []string) error
	TakeGroupByID(ctx context.Context, groupID string) (group *relation.Group, err error)

	//mongo
	GetSuperGroupByID(ctx context.Context, groupID string) (superGroup *unrelation.SuperGroup, err error)
}

type GroupController struct {
	database DataBase
}

func NewGroupController(db *gorm.DB, rdb redis.UniversalClient, mgoDB *mongo.Database) GroupInterface {
	groupController := &GroupController{database: newGroupDatabase(db, rdb, mgoDB)}
	return groupController
}

func (g *GroupController) FindGroupsByID(ctx context.Context, groupIDs []string) (groups []*relation.Group, err error) {
	return g.database.FindGroupsByID(ctx, groupIDs)
}

func (g *GroupController) CreateGroup(ctx context.Context, groups []*relation.Group) error {
	return g.database.CreateGroup(ctx, groups)
}

func (g *GroupController) DeleteGroupByIDs(ctx context.Context, groupIDs []string) error {
	return g.database.DeleteGroupByIDs(ctx, groupIDs)
}

func (g *GroupController) TakeGroupByID(ctx context.Context, groupID string) (group *relation.Group, err error) {
	return g.database.TakeGroupByID(ctx, groupID)
}

func (g *GroupController) GetSuperGroupByID(ctx context.Context, groupID string) (superGroup *unrelation.SuperGroup, err error) {
	return g.database.GetSuperGroupByID(ctx, groupID)
}

type DataBase interface {
	FindGroupsByID(ctx context.Context, groupIDs []string) (groups []*relation.Group, err error)
	CreateGroup(ctx context.Context, groups []*relation.Group) error
	DeleteGroupByIDs(ctx context.Context, groupIDs []string) error
	TakeGroupByID(ctx context.Context, groupID string) (group *relation.Group, err error)
	GetSuperGroupByID(ctx context.Context, groupID string) (superGroup *unrelation.SuperGroup, err error)
}

type GroupDataBase struct {
	groupDB       *relation.Group
	groupMemberDB *relation.GroupMember
	cache         *cache.GroupCache
	db            *gorm.DB
	mongoDB       *unrelation.SuperGroupMgoDB
}

func newGroupDatabase(db *gorm.DB, rdb redis.UniversalClient, mgoDB *mongo.Database) DataBase {
	group := relation.NewGroupDB(db)
	groupMember := relation.NewGroupMemberDB(db)
	database := &GroupDataBase{
		groupDB:       group,
		groupMemberDB: groupMember,
		db:            db,
		cache: cache.NewGroupCache(rdb, group, rockscache.Options{
			RandomExpireAdjustment: 0.2,
			DisableCacheRead:       false,
			DisableCacheDelete:     false,
			StrongConsistency:      true,
		}),
		mongoDB: unrelation.NewSuperGroupMgoDB(mgoDB),
	}
	return database
}

func (g *GroupDataBase) FindGroupsByID(ctx context.Context, groupIDs []string) (groups []*relation.Group, err error) {
	return g.cache.GetGroupsInfoFromCache(ctx, groupIDs)
}

func (g *GroupDataBase) CreateGroup(ctx context.Context, groups []*relation.Group) error {
	return g.groupDB.Create(ctx, groups)
}

func (g *GroupDataBase) DeleteGroupByIDs(ctx context.Context, groupIDs []string) error {
	return g.groupDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := g.groupDB.Delete(ctx, groupIDs, tx); err != nil {
			return err
		}
		if err := g.cache.DelGroupsInfoFromCache(ctx, groupIDs); err != nil {
			return err
		}
		return nil
	})
}

func (g *GroupDataBase) TakeGroupByID(ctx context.Context, groupID string) (group *relation.Group, err error) {
	return g.cache.GetGroupInfoFromCache(ctx, groupID)
}

func (g *GroupDataBase) Update(ctx context.Context, groups []*relation.Group) error {
	return g.groupDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := g.groupDB.Update(ctx, groups, tx); err != nil {
			return err
		}
		var groupIDs []string
		for _, group := range groups {
			groupIDs = append(groupIDs, group.GroupID)
		}
		if err := g.cache.DelGroupsInfoFromCache(ctx, groupIDs); err != nil {
			return err
		}
		return nil
	})
}

func (g *GroupDataBase) GetSuperGroupByID(ctx context.Context, groupID string) (superGroup *unrelation.SuperGroup, err error) {
	return g.mongoDB.GetSuperGroup(ctx, groupID)
}

func (g *GroupDataBase) CreateGroupAndMember(ctx context.Context, groups []*relation.Group, groupMember []*relation.GroupMember) error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		if err := g.groupDB.Create(ctx, groups, tx); err != nil {
			return err
		}
		if err := g.groupMemberDB.Create(ctx, groupMember, tx); err != nil {
			return err
		}
		return nil
	})
}
