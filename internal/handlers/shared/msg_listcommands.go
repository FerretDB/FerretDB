// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shared

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListCommands provides a list of all database commands.
func (h *Handler) MsgListCommands(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg
	err := reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"commands", supportedCommandList(),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

func supportedCommandList() []types.Document {
	return []types.Document{
		types.MustMakeDocument(
			"buildInfo", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"count", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"create", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"createIndexes", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"delete", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"drop", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropDatabase", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"find", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getCmdLineOpts", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getLog", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getParameter", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"hello", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"hostInfo", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"insert", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"isMaster", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"listCollections", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"listCommands", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"listDatabases", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"ping", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"serverStatus", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"update", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"whatsmyuri", types.MustMakeDocument(
				"help", "",
			),
		),
	}
}

func totalCommandList() []types.Document {
	return []types.Document{
		types.MustMakeDocument(
			"abortTransaction", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"aggregate", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"appendOplogNote", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"applyOps", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"authenticate", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"autoSplitVector", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"availableQueryOptions", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"buildInfo", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"checkShardingIndex", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"cleanupOrphaned", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"cloneCollectionAsCapped", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"collMod", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"collStats", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"commitTransaction", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"compact", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"connPoolStats", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"connPoolSync", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"connectionStatus", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"convertToCapped", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"coordinateCommitTransaction", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"count", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"create", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"createIndexes", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"createRole", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"createUser", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"currentOp", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dataSize", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dbHash", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dbStats", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"delete", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"distinct", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"donorAbortMigration", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"donorForgetMigration", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"donorStartMigration", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"driverOIDTest", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"drop", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropAllRolesFromDatabase", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropAllUsersFromDatabase", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropConnections", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropDatabase", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropIndexes", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropRole", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"dropUser", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"endSessions", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"explain", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"features", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"filemd5", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"find", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"findAndModify", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"flushRouterConfig", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"fsync", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"fsyncUnlock", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getCmdLineOpts", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getDatabaseVersion", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getDefaultRWConcern", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getDiagnosticData", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getFreeMonitoringStatus", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getLastError", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getLog", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getMore", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getParameter", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getShardMap", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getShardVersion", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"getnonce", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"grantPrivilegesToRole", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"grantRolesToRole", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"grantRolesToUser", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"hello", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"hostInfo", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"insert", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"internalRenameIfOptionsAndIndexesMatch", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"invalidateUserCache", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"isMaster", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"killAllSessions", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"killAllSessionsByPattern", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"killCursors", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"killOp", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"killSessions", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"listCollections", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"listCommands", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"listDatabases", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"listIndexes", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"lockInfo", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"logRotate", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"logout", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"mapReduce", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"mergeChunks", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"moveChunk", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"ping", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"planCacheClear", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"planCacheClearFilters", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"planCacheListFilters", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"planCacheSetFilter", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"prepareTransaction", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"profile", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"reIndex", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"recipientForgetMigration", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"recipientSyncData", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"refreshSessions", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"renameCollection", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"repairDatabase", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetAbortPrimaryCatchUp", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetFreeze", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetGetConfig", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetGetRBID", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetGetStatus", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetHeartbeat", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetInitiate", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetMaintenance", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetReconfig", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetRequestVotes", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetResizeOplog", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetStepDown", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetStepUp", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetSyncFrom", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"replSetUpdatePosition", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"revokePrivilegesFromRole", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"revokeRolesFromRole", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"revokeRolesFromUser", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"rolesInfo", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"rotateCertificates", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"saslContinue", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"saslStart", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"serverStatus", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"setDefaultRWConcern", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"setFeatureCompatibilityVersion", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"setFreeMonitoring", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"setIndexCommitQuorum", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"setParameter", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"setShardVersion", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"shardingState", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"shutdown", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"splitChunk", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"splitVector", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"startRecordingTraffic", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"startSession", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"stopRecordingTraffic", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"top", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"update", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"updateRole", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"updateUser", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"usersInfo", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"validate", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"validateDBMetadata", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"voteCommitIndexBuild", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"waitForFailPoint", types.MustMakeDocument(
				"help", "",
			),
		),
		types.MustMakeDocument(
			"whatsmyuri", types.MustMakeDocument(
				"help", "",
			),
		),
	}
}
