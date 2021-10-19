// +build wireinject
// The build tag makes sure the stub is not built in the final build.

package business

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	commonadapter "github.com/ApsaraDB/PolarDB-Stack-Common/business/adapter"
	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	commonservice "github.com/ApsaraDB/PolarDB-Stack-Common/business/service"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/adapter"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/service"
)

var commonSet = wire.NewSet(
	wire.Bind(new(commondomain.IClassQuery), new(*commonadapter.ClassQuery)),
	wire.Bind(new(commondomain.IEngineParamsClassQuery), new(*commonadapter.EngineParamsClassQuery)),
	wire.Bind(new(commondomain.IEngineParamsRepository), new(*commonadapter.EngineParamsRepository)),
	wire.Bind(new(commondomain.IEngineParamsTemplateQuery), new(*commonadapter.EngineParamsTemplateQuery)),
	wire.Bind(new(commondomain.IIdGenerator), new(*commonadapter.IdGenerator)),
	wire.Bind(new(commondomain.IPortGenerator), new(*commonadapter.PortGenerator)),
	wire.Bind(new(commondomain.IMinorVersionQuery), new(*commonadapter.MinorVersionQuery)),
	wire.Bind(new(commondomain.IAccountRepository), new(*commonadapter.AccountRepository)),
	wire.Bind(new(commondomain.IManagerClient), new(*commonadapter.ManagerClient)),
	wire.Bind(new(commondomain.IStorageManager), new(*commonadapter.StorageManager)),
	wire.Bind(new(commondomain.IPfsdToolClient), new(*commonadapter.PfsdToolClient)),
	wire.Bind(new(commondomain.IClusterManagerClient), new(*commonadapter.ClusterManagerClient)),
	wire.Bind(new(commondomain.IClusterManagerCreator), new(*commonadapter.ClusterManagerCreator)),
	wire.Bind(new(commondomain.IClusterManagerRemover), new(*commonadapter.ClusterManagerRemover)),
	adapter.NewMdpAccountRepository,
	adapter.NewMpdEngineParamsRepository,
	commonadapter.NewClusterManagerRemover,
	commonadapter.NewClusterManagerCreator,
	commonadapter.NewClassQuery,
	commonadapter.NewEngineParamsClassQuery,
	commonadapter.NewEngineParamsTemplateQuery,
	commonadapter.NewIdGenerator,
	commonadapter.NewPortGenerator,
	commonadapter.NewMinorVersionQuery,
	commonadapter.NewManagerClient,
	commonadapter.NewStorageManager,
	commonadapter.NewPfsdToolClient,
	commonadapter.NewClusterManagerClient,
	service.NewShardStorageClusterService,
	service.NewLocalStorageClusterService,
	commonservice.NewEngineParamsTemplateService,
	commonservice.NewEngineClassService,
	commonservice.NewMinorVersionService,
	commonservice.NewCmCreatorService,
)

var sharedStorageSet = wire.NewSet(
	wire.Bind(new(domain.ISharedStorageClusterRepository), new(*adapter.SharedStorageClusterRepository)),
	wire.Bind(new(commonadapter.IEnvGetStrategy), new(*adapter.SharedStorageClusterEnvGetStrategy)),
	wire.Bind(new(commondomain.IEnginePodManager), new(*adapter.SharedStoragePodManager)),
	adapter.NewSharedStorageClusterRepository,
	adapter.NewSharedStorageClusterEnvGetStrategy,
	adapter.NewSharedStoragePodManager,
)

var localStorageSet = wire.NewSet(
	wire.Bind(new(domain.ILocalStorageClusterRepository), new(*adapter.LocalStorageClusterRepository)),
	wire.Bind(new(commonadapter.IEnvGetStrategy), new(*adapter.LocalStorageClusterEnvGetStrategy)),
	wire.Bind(new(commondomain.IEnginePodManager), new(*adapter.LocalStoragePodManager)),
	adapter.NewLocalStorageClusterRepository,
	adapter.NewLocalStorageClusterEnvGetStrategy,
	adapter.NewLocalStoragePodManager,
)

func NewSharedStorageClusterService(logger logr.Logger) *service.SharedStorageClusterService {
	wire.Build(commonSet, sharedStorageSet)
	return nil
}

func NewLocalStorageClusterService(logger logr.Logger) *service.LocalStorageClusterService {
	wire.Build(commonSet, localStorageSet)
	return nil
}

func NewEngineParamsTemplateService(logger logr.Logger) *commonservice.EngineParamsTemplateService {
	wire.Build(commonSet)
	return nil
}

func NewEngineClassService(logger logr.Logger) *commonservice.EngineClassService {
	wire.Build(commonSet)
	return nil
}

func NewMinorVersionService(logger logr.Logger) *commonservice.MinorVersionService {
	wire.Build(commonSet)
	return nil
}

func NewCmCreatorService(logger logr.Logger) *commonservice.CmCreatorService {
	wire.Build(commonSet)
	return nil
}
