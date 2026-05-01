package mac

import (
	"sync"

	"github.com/cruciblehq/crux/internal/runtime/shared/macspec"
)

var (
	catalogOnce sync.Once         // Ensures catalog is built once and reused thereafter.
	catalogReg  *macspec.Registry // The catalog registry, populated on first call to catalog
)

// Returns the catalog populated with all LSM hooks declared in kernel 6.18.
//
// The registry is built once on first call and reused thereafter; callers must
// treat it as immutable. The registry is derived from the kernel's source code
// at include/linux/lsm_hook_defs.h.
func catalog() *macspec.Registry {
	catalogOnce.Do(func() {
		r := macspec.NewRegistry()
		registerAttrHooks(r)
		registerAuditHooks(r)
		registerBdevHooks(r)
		registerBinderHooks(r)
		registerBpfHooks(r)
		registerBprmHooks(r)
		registerCapHooks(r)
		registerCredHooks(r)
		registerFileHooks(r)
		registerFsHooks(r)
		registerIbHooks(r)
		registerInodeHooks(r)
		registerIpcHooks(r)
		registerKernelHooks(r)
		registerKeyHooks(r)
		registerLockdownHooks(r)
		registerMiscHooks(r)
		registerMmHooks(r)
		registerNetHooks(r)
		registerPathHooks(r)
		registerPerfHooks(r)
		registerPtraceHooks(r)
		registerQuotaHooks(r)
		registerSbHooks(r)
		registerSyslogHooks(r)
		registerTaskHooks(r)
		registerTimeHooks(r)
		registerUringHooks(r)
		registerWatchHooks(r)
		registerXfrmHooks(r)
		catalogReg = r
	})
	return catalogReg
}

// Returns a fields map keyed by field name from the given Field values.
func fieldsOf(fs ...macspec.Field) map[string]macspec.Field {
	out := make(map[string]macspec.Field, len(fs))
	for _, f := range fs {
		out[f.Name] = f
	}
	return out
}

// Registers all hooks in the attr family.
func registerAttrHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookGetselfattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldAttr, fieldSize, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookSetselfattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldAttr, fieldSize, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookGetprocattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldName)})
	r.AddHook(macspec.Hook{Name: hookSetprocattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldName, fieldSize)})
	r.AddHook(macspec.Hook{Name: hookIsmaclabel, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldName)})
	r.AddHook(macspec.Hook{Name: hookSecidToSecctx, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookLsmpropToSecctx, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookSecctxToSecid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSeclen, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookReleaseSecctx, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
}

// Registers all hooks in the audit family.
func registerAuditHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookAuditRuleInit, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldField, fieldOp, fieldRulestr, fieldGfp)})
	r.AddHook(macspec.Hook{Name: hookAuditRuleKnown, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookAuditRuleMatch, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldField, fieldOp)})
	r.AddHook(macspec.Hook{Name: hookAuditRuleFree, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
}

// Registers all hooks in the bdev family.
func registerBdevHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookBdevAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBdevDev, fieldBdevPath)})
	r.AddHook(macspec.Hook{Name: hookBdevFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBdevDev, fieldBdevPath)})
	r.AddHook(macspec.Hook{Name: hookBdevSetintegrity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBdevDev, fieldBdevPath, fieldType, fieldSize)})
}

// Registers all hooks in the binder family.
func registerBinderHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookBinderSetContextMgr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookBinderTransaction, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookBinderTransferBinder, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookBinderTransferFile, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
}

// Registers all hooks in the bpf family.
func registerBpfHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookBpf, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCmd, fieldSize, fieldKernel)})
	r.AddHook(macspec.Hook{Name: hookBpfMap, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfMapType, fieldBpfMapId, fieldBpfFmode, fieldFmode)})
	r.AddHook(macspec.Hook{Name: hookBpfProg, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfProgType, fieldBpfProgId)})
	r.AddHook(macspec.Hook{Name: hookBpfMapCreate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfMapType, fieldBpfMapId, fieldBpfFmode, fieldBpfTokenKind, fieldKernel)})
	r.AddHook(macspec.Hook{Name: hookBpfMapFree, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfMapType, fieldBpfMapId, fieldBpfFmode)})
	r.AddHook(macspec.Hook{Name: hookBpfProgLoad, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfProgType, fieldBpfProgId, fieldBpfTokenKind, fieldKernel)})
	r.AddHook(macspec.Hook{Name: hookBpfProgFree, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfProgType, fieldBpfProgId)})
	r.AddHook(macspec.Hook{Name: hookBpfTokenCreate, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfTokenKind, fieldFilePath, fieldFileIno, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookBpfTokenFree, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfTokenKind)})
	r.AddHook(macspec.Hook{Name: hookBpfTokenCmd, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfTokenKind, fieldCmd)})
	r.AddHook(macspec.Hook{Name: hookBpfTokenCapable, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBpfTokenKind, fieldCap)})
}

// Registers all hooks in the bprm family.
func registerBprmHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookBprmCredsForExec, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBprmFilename, fieldBprmInterp, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookBprmCredsFromFile, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBprmFilename, fieldBprmInterp, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookBprmCheckSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBprmFilename, fieldBprmInterp, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookBprmCommittingCreds, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBprmFilename, fieldBprmInterp, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookBprmCommittedCreds, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldBprmFilename, fieldBprmInterp, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
}

// Registers all hooks in the cap family.
func registerCapHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookCapget, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookCapset, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookCapable, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldNsUsernsId, fieldCap, fieldOpts)})
}

// Registers all hooks in the cred family.
func registerCredHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookCredAllocBlank, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldGfp)})
	r.AddHook(macspec.Hook{Name: hookCredFree, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookCredPrepare, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldGfp)})
	r.AddHook(macspec.Hook{Name: hookCredTransfer, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookCredGetsecid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookCredGetlsmprop, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
}

// Registers all hooks in the file family.
func registerFileHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookFilePermission, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMask)})
	r.AddHook(macspec.Hook{Name: hookFileAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookFileRelease, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookFileFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookFileIoctl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldCmd, fieldArg)})
	r.AddHook(macspec.Hook{Name: hookFileIoctlCompat, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldCmd, fieldArg)})
	r.AddHook(macspec.Hook{Name: hookMmapAddr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldAddr)})
	r.AddHook(macspec.Hook{Name: hookMmapFile, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldReqprot, fieldProt, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookFileMprotect, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldVmaStart, fieldVmaEnd, fieldVmaFlags, fieldReqprot, fieldProt)})
	r.AddHook(macspec.Hook{Name: hookFileLock, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldCmd)})
	r.AddHook(macspec.Hook{Name: hookFileFcntl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldCmd, fieldArg)})
	r.AddHook(macspec.Hook{Name: hookFileSetFowner, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookFileSendSigiotask, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldFownPid, fieldFownUid, fieldSig)})
	r.AddHook(macspec.Hook{Name: hookFileReceive, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookFileOpen, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookFilePostOpen, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMask)})
	r.AddHook(macspec.Hook{Name: hookFileTruncate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
}

// Registers all hooks in the fs family.
func registerFsHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookFsContextSubmount, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFcFstype, fieldFcParamName, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookFsContextDup, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFcFstype, fieldFcParamName)})
	r.AddHook(macspec.Hook{Name: hookFsContextParseParam, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFcFstype, fieldFcParamName)})
}

// Registers all hooks in the ib family.
func registerIbHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookIbPkeyAccess, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSubnetPrefix, fieldPkey)})
	r.AddHook(macspec.Hook{Name: hookIbEndportManageSubnet, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDevName, fieldPortNum)})
	r.AddHook(macspec.Hook{Name: hookIbAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
}

// Registers all hooks in the inode family.
func registerInodeHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookInodeAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookInodeFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookInodeFreeSecurityRcu, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookInodeInitSecurity, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldDirPath, fieldDirIno, fieldDirDev, fieldName, fieldXattrName, fieldXattrValue, fieldXattrCount)})
	r.AddHook(macspec.Hook{Name: hookInodeInitSecurityAnon, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldName)})
	r.AddHook(macspec.Hook{Name: hookInodeCreate, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMode)})
	r.AddHook(macspec.Hook{Name: hookInodePostCreateTmpfile, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookInodeLink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldOldPath, fieldOldIno, fieldOldDev, fieldOldDirPath, fieldDirPath, fieldDirIno, fieldDirDev, fieldNewPath, fieldNewDirPath)})
	r.AddHook(macspec.Hook{Name: hookInodeUnlink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeSymlink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldOldName)})
	r.AddHook(macspec.Hook{Name: hookInodeMkdir, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMode)})
	r.AddHook(macspec.Hook{Name: hookInodeRmdir, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeMknod, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMode, fieldDev)})
	r.AddHook(macspec.Hook{Name: hookInodeRename, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldOldDirPath, fieldOldDirIno, fieldOldPath, fieldOldIno, fieldOldDev, fieldNewDirPath, fieldNewDirIno, fieldNewPath)})
	r.AddHook(macspec.Hook{Name: hookInodeReadlink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeFollowLink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldRcu)})
	r.AddHook(macspec.Hook{Name: hookInodePermission, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldMask)})
	r.AddHook(macspec.Hook{Name: hookInodeSetattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldIattrMask, fieldIattrMode, fieldIattrUid, fieldIattrGid)})
	r.AddHook(macspec.Hook{Name: hookInodePostSetattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldIaValid)})
	r.AddHook(macspec.Hook{Name: hookInodeGetattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookInodeXattrSkipcap, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldName)})
	r.AddHook(macspec.Hook{Name: hookInodeSetxattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldName, fieldSize, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookInodePostSetxattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldName, fieldSize, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookInodeGetxattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldName)})
	r.AddHook(macspec.Hook{Name: hookInodeListxattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeRemovexattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldName)})
	r.AddHook(macspec.Hook{Name: hookInodePostRemovexattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldName)})
	r.AddHook(macspec.Hook{Name: hookInodeFileSetattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldFkattrMask, fieldFkattrFlags)})
	r.AddHook(macspec.Hook{Name: hookInodeFileGetattr, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldFkattrMask, fieldFkattrFlags)})
	r.AddHook(macspec.Hook{Name: hookInodeSetAcl, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldAclType)})
	r.AddHook(macspec.Hook{Name: hookInodePostSetAcl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldAclType)})
	r.AddHook(macspec.Hook{Name: hookInodeGetAcl, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeRemoveAcl, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodePostRemoveAcl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeNeedKillpriv, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeKillpriv, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeGetsecurity, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldName, fieldAlloc)})
	r.AddHook(macspec.Hook{Name: hookInodeSetsecurity, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldName, fieldSize, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookInodeListsecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldBufferSize)})
	r.AddHook(macspec.Hook{Name: hookInodeGetlsmprop, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookInodeCopyUp, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookInodeCopyUpXattr, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldName)})
	r.AddHook(macspec.Hook{Name: hookInodeSetintegrity, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldType, fieldSize)})
	r.AddHook(macspec.Hook{Name: hookKernfsInitSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldKernfsId, fieldKernfsName)})
	r.AddHook(macspec.Hook{Name: hookDInstantiate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookInodeInvalidateSecctx, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookInodeNotifysecctx, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev, fieldCtxlen)})
	r.AddHook(macspec.Hook{Name: hookInodeSetsecctx, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldCtxlen)})
	r.AddHook(macspec.Hook{Name: hookInodeGetsecctx, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
}

// Registers all hooks in the ipc family.
func registerIpcHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookIpcPermission, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldFlag)})
	r.AddHook(macspec.Hook{Name: hookIpcGetlsmprop, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode)})
	r.AddHook(macspec.Hook{Name: hookMsgMsgAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldMsgType, fieldMsgSize)})
	r.AddHook(macspec.Hook{Name: hookMsgMsgFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldMsgType, fieldMsgSize)})
	r.AddHook(macspec.Hook{Name: hookMsgQueueAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode)})
	r.AddHook(macspec.Hook{Name: hookMsgQueueFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode)})
	r.AddHook(macspec.Hook{Name: hookMsgQueueAssociate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldMsqflg)})
	r.AddHook(macspec.Hook{Name: hookMsgQueueMsgctl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldCmd)})
	r.AddHook(macspec.Hook{Name: hookMsgQueueMsgsnd, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldMsgType, fieldMsgSize, fieldMsqflg)})
	r.AddHook(macspec.Hook{Name: hookMsgQueueMsgrcv, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldMsgType, fieldMsgSize, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldType, fieldMode)})
	r.AddHook(macspec.Hook{Name: hookShmAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode)})
	r.AddHook(macspec.Hook{Name: hookShmFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode)})
	r.AddHook(macspec.Hook{Name: hookShmAssociate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldShmflg)})
	r.AddHook(macspec.Hook{Name: hookShmShmctl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldCmd)})
	r.AddHook(macspec.Hook{Name: hookShmShmat, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldShmaddr, fieldShmflg)})
	r.AddHook(macspec.Hook{Name: hookSemAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode)})
	r.AddHook(macspec.Hook{Name: hookSemFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode)})
	r.AddHook(macspec.Hook{Name: hookSemAssociate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldSemflg)})
	r.AddHook(macspec.Hook{Name: hookSemSemctl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldCmd)})
	r.AddHook(macspec.Hook{Name: hookSemSemop, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldIpcId, fieldIpcKey, fieldIpcUid, fieldIpcGid, fieldIpcMode, fieldNsops, fieldAlter)})
}

// Registers all hooks in the kernel family.
func registerKernelHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookKernelActAs, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookKernelCreateFilesAs, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookKernelModuleRequest, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldKmodName)})
	r.AddHook(macspec.Hook{Name: hookKernelLoadData, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldId, fieldContents)})
	r.AddHook(macspec.Hook{Name: hookKernelPostLoadData, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSize, fieldId, fieldDescription)})
	r.AddHook(macspec.Hook{Name: hookKernelReadFile, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldId, fieldContents)})
	r.AddHook(macspec.Hook{Name: hookKernelPostReadFile, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldSize, fieldId)})
}

// Registers all hooks in the key family.
func registerKeyHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookKeyAlloc, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldKeySerial, fieldKeyType, fieldKeyUid, fieldKeyGid, fieldKeyFlags, fieldCredUid, fieldCredGid, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookKeyPermission, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldNeedPerm)})
	r.AddHook(macspec.Hook{Name: hookKeyGetsecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldKeySerial, fieldKeyType, fieldKeyUid, fieldKeyGid, fieldKeyFlags)})
	r.AddHook(macspec.Hook{Name: hookKeyPostCreateOrUpdate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldKeySerial, fieldKeyType, fieldKeyUid, fieldKeyGid, fieldKeyFlags, fieldPayloadLen, fieldFlags, fieldCreate)})
}

// Registers all hooks in the lockdown family.
func registerLockdownHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookLockedDown, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldLsmflag)})
}

// Registers all hooks in the misc family.
func registerMiscHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookInitramfsPopulated, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
}

// Registers all hooks in the mm family.
func registerMmHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookVmEnoughMemory, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldPages)})
}

// Registers all hooks in the net family.
func registerNetHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookNetlinkSend, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSkbProto, fieldSkbLen)})
	r.AddHook(macspec.Hook{Name: hookUnixStreamConnect, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookUnixMaySend, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSocketCreate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFamily, fieldType, fieldProtocol, fieldKern)})
	r.AddHook(macspec.Hook{Name: hookSocketPostCreate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldFamily, fieldType, fieldProtocol, fieldKern)})
	r.AddHook(macspec.Hook{Name: hookSocketSocketpair, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSocketBind, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldAddrFamily, fieldAddrPort, fieldAddrAddr, fieldAddrlen)})
	r.AddHook(macspec.Hook{Name: hookSocketConnect, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldAddrFamily, fieldAddrPort, fieldAddrAddr, fieldAddrlen)})
	r.AddHook(macspec.Hook{Name: hookSocketListen, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldBacklog)})
	r.AddHook(macspec.Hook{Name: hookSocketAccept, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSocketSendmsg, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSize)})
	r.AddHook(macspec.Hook{Name: hookSocketRecvmsg, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSize, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookSocketGetsockname, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSocketGetpeername, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSocketGetsockopt, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldLevel, fieldOptname)})
	r.AddHook(macspec.Hook{Name: hookSocketSetsockopt, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldLevel, fieldOptname)})
	r.AddHook(macspec.Hook{Name: hookSocketShutdown, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldHow)})
	r.AddHook(macspec.Hook{Name: hookSocketSockRcvSkb, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSkbProto, fieldSkbLen)})
	r.AddHook(macspec.Hook{Name: hookSocketGetpeersecStream, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldLen)})
	r.AddHook(macspec.Hook{Name: hookSocketGetpeersecDgram, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSkbProto, fieldSkbLen, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookSkAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldFamily, fieldPriority)})
	r.AddHook(macspec.Hook{Name: hookSkFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSkCloneSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSkGetsecid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookSockGraft, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookInetConnRequest, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSkbProto, fieldSkbLen, fieldReqFamily)})
	r.AddHook(macspec.Hook{Name: hookInetCskClone, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldReqFamily)})
	r.AddHook(macspec.Hook{Name: hookInetConnEstablished, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldSkbProto, fieldSkbLen)})
	r.AddHook(macspec.Hook{Name: hookSecmarkRelabelPacket, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookSecmarkRefcountInc, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookSecmarkRefcountDec, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookReqClassifyFlow, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldReqFamily)})
	r.AddHook(macspec.Hook{Name: hookTunDevAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTunDevCreate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTunDevAttachQueue, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTunDevAttach, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookTunDevOpen, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookSctpAssocRequest, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSctpAssocId, fieldSkbProto, fieldSkbLen)})
	r.AddHook(macspec.Hook{Name: hookSctpBindConnect, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport, fieldOptname, fieldAddrFamily, fieldAddrPort, fieldAddrAddr, fieldAddrlen)})
	r.AddHook(macspec.Hook{Name: hookSctpSkClone, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSctpAssocId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
	r.AddHook(macspec.Hook{Name: hookSctpAssocEstablished, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSctpAssocId, fieldSkbProto, fieldSkbLen)})
	r.AddHook(macspec.Hook{Name: hookMptcpAddSubflow, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSockFamily, fieldSockType, fieldSockProtocol, fieldSockSaddr, fieldSockSport, fieldSockDaddr, fieldSockDport)})
}

// Registers all hooks in the path family.
func registerPathHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookPathUnlink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookPathMkdir, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMode)})
	r.AddHook(macspec.Hook{Name: hookPathRmdir, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookPathMknod, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMode, fieldDev)})
	r.AddHook(macspec.Hook{Name: hookPathPostMknod, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookPathTruncate, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookPathSymlink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDirPath, fieldDirIno, fieldDirDev, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldOldName)})
	r.AddHook(macspec.Hook{Name: hookPathLink, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldOldPath, fieldOldIno, fieldOldDev, fieldOldDirPath, fieldFilePath, fieldFileIno, fieldFileDev, fieldNewPath, fieldNewDirPath)})
	r.AddHook(macspec.Hook{Name: hookPathRename, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldOldPath, fieldOldIno, fieldOldDev, fieldOldDirPath, fieldNewPath, fieldNewDirPath, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookPathChmod, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldMode)})
	r.AddHook(macspec.Hook{Name: hookPathChown, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldUid, fieldGid)})
	r.AddHook(macspec.Hook{Name: hookPathChroot, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookPathNotify, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldMask, fieldObjType)})
}

// Registers all hooks in the perf family.
func registerPerfHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookPerfEventOpen, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldType)})
	r.AddHook(macspec.Hook{Name: hookPerfEventAlloc, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldPerfType, fieldPerfConfig)})
	r.AddHook(macspec.Hook{Name: hookPerfEventRead, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldPerfType, fieldPerfConfig)})
	r.AddHook(macspec.Hook{Name: hookPerfEventWrite, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldPerfType, fieldPerfConfig)})
}

// Registers all hooks in the ptrace family.
func registerPtraceHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookPtraceAccessCheck, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldMode)})
	r.AddHook(macspec.Hook{Name: hookPtraceTraceme, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
}

// Registers all hooks in the quota family.
func registerQuotaHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookQuotactl, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCmds, fieldType, fieldId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookQuotaOn, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
}

// Registers all hooks in the sb family.
func registerSbHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookSbAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookSbDelete, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookSbFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookSbFreeMntOpts, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookSbEatLsmOpts, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookSbMntOptsCompat, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookSbRemount, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookSbKernMount, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookSbShowOptions, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype)})
	r.AddHook(macspec.Hook{Name: hookSbStatfs, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash)})
	r.AddHook(macspec.Hook{Name: hookSbMount, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldDevName, fieldFilePath, fieldFileIno, fieldFileDev, fieldType, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookSbUmount, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldMntDevname, fieldMntFstype, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookSbPivotroot, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldOldPath, fieldOldIno, fieldOldDev, fieldOldDirPath, fieldNewPath, fieldNewDirPath)})
	r.AddHook(macspec.Hook{Name: hookSbSetMntOpts, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype, fieldKernFlags, fieldSetKernFlags)})
	r.AddHook(macspec.Hook{Name: hookSbCloneMntOpts, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSbDev, fieldSbFstype, fieldKernFlags, fieldSetKernFlags)})
	r.AddHook(macspec.Hook{Name: hookMoveMount, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldOldPath, fieldOldIno, fieldOldDev, fieldOldDirPath, fieldNewPath, fieldNewDirPath)})
	r.AddHook(macspec.Hook{Name: hookDentryInitSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMode, fieldName)})
	r.AddHook(macspec.Hook{Name: hookDentryCreateFilesAs, Sleepable: true, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldFilePath, fieldFileIno, fieldFileDev, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileImaHash, fieldMode, fieldName, fieldCredUid, fieldCredGid)})
}

// Registers all hooks in the syslog family.
func registerSyslogHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookSyslog, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldType)})
}

// Registers all hooks in the task family.
func registerTaskHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookTaskAlloc, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldCloneFlags)})
	r.AddHook(macspec.Hook{Name: hookTaskFree, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskFixSetuid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookTaskFixSetgid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookTaskFixSetgroups, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookTaskSetpgid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldPgid)})
	r.AddHook(macspec.Hook{Name: hookTaskGetpgid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskGetsid, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookCurrentGetlsmpropSubj, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskGetlsmpropObj, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskSetnice, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldNice)})
	r.AddHook(macspec.Hook{Name: hookTaskSetioprio, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldIoprio)})
	r.AddHook(macspec.Hook{Name: hookTaskGetioprio, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskPrlimit, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldFlags)})
	r.AddHook(macspec.Hook{Name: hookTaskSetrlimit, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldResource, fieldRlimCur, fieldRlimMax)})
	r.AddHook(macspec.Hook{Name: hookTaskSetscheduler, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskGetscheduler, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskMovememory, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId)})
	r.AddHook(macspec.Hook{Name: hookTaskKill, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldSiginfoSigno, fieldSiginfoCode, fieldSiginfoUid, fieldSig, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookTaskPrctl, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldOption, fieldArg2, fieldArg3, fieldArg4, fieldArg5)})
	r.AddHook(macspec.Hook{Name: hookTaskToInode, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTargetUid, fieldTargetGid, fieldTargetPid, fieldTargetTgid, fieldTargetCgroupId, fieldFileIno, fieldFileMode, fieldFileUid, fieldFileGid, fieldFileDev)})
	r.AddHook(macspec.Hook{Name: hookUsernsCreate, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
}

// Registers all hooks in the time family.
func registerTimeHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookSettime, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldTimeSec, fieldTimeNsec, fieldTimeMinuteswest)})
}

// Registers all hooks in the uring family.
func registerUringHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookUringOverrideCreds, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid)})
	r.AddHook(macspec.Hook{Name: hookUringSqpoll, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
	r.AddHook(macspec.Hook{Name: hookUringCmd, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldUringCmdOp, fieldUringFlags)})
	r.AddHook(macspec.Hook{Name: hookUringAllowed, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId)})
}

// Registers all hooks in the watch family.
func registerWatchHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookPostNotification, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldCredUid, fieldCredGid, fieldWatchType, fieldWatchSubtype)})
	r.AddHook(macspec.Hook{Name: hookWatchKey, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldKeySerial, fieldKeyType, fieldKeyUid, fieldKeyGid, fieldKeyFlags)})
}

// Registers all hooks in the xfrm family.
func registerXfrmHooks(r *macspec.Registry) {
	r.AddHook(macspec.Hook{Name: hookXfrmPolicyAllocSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid, fieldGfp)})
	r.AddHook(macspec.Hook{Name: hookXfrmPolicyCloneSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid)})
	r.AddHook(macspec.Hook{Name: hookXfrmPolicyFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid)})
	r.AddHook(macspec.Hook{Name: hookXfrmPolicyDeleteSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid)})
	r.AddHook(macspec.Hook{Name: hookXfrmStateAlloc, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid)})
	r.AddHook(macspec.Hook{Name: hookXfrmStateAllocAcquire, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid, fieldSecid)})
	r.AddHook(macspec.Hook{Name: hookXfrmStateFreeSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid)})
	r.AddHook(macspec.Hook{Name: hookXfrmStateDeleteSecurity, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid)})
	r.AddHook(macspec.Hook{Name: hookXfrmPolicyLookup, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid, fieldFlSecid)})
	r.AddHook(macspec.Hook{Name: hookXfrmStatePolFlowMatch, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldXfrmSpi, fieldXfrmProto, fieldXfrmReqid)})
	r.AddHook(macspec.Hook{Name: hookXfrmDecodeSession, Sleepable: false, Fields: fieldsOf(fieldTaskUid, fieldTaskGid, fieldTaskPid, fieldTaskTgid, fieldTaskComm, fieldTaskCgroupId, fieldSkbProto, fieldSkbLen, fieldSecid, fieldCkall)})
}
