package purger

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/siddontang/go/sync2"

	"github.com/pingcap/dm/pkg/streamer"
	"github.com/pingcap/dm/pkg/utils"
)

var (
	fakeTaskName = strategyFilename.String()
)

// filenameArgs represents args needed by filenameStrategy
// NOTE: should handle master-slave switch
type filenameArgs struct {
	relayBaseDir string
	filename     string // specified end safe filename
	subDir       string // sub dir for @filename, empty indicates latest sub dir
	uuids        []string
	safeRelayLog *streamer.RelayLogInfo // all relay log files prior to this should be purged
}

func (fa *filenameArgs) SetActiveRelayLog(active *streamer.RelayLogInfo) {
	uuid := fa.subDir
	if len(uuid) == 0 && len(fa.uuids) > 0 {
		// no sub dir specified, use the latest one
		uuid = fa.uuids[len(fa.uuids)-1]
	}
	_, endSuffix, _ := utils.ParseSuffixForUUID(uuid)

	safeRelayLog := &streamer.RelayLogInfo{
		TaskName:   fakeTaskName,
		UUID:       uuid,
		UUIDSuffix: endSuffix,
		Filename:   fa.filename,
	}

	if active.Earlier(safeRelayLog) {
		safeRelayLog = active
	}

	fa.safeRelayLog = safeRelayLog

	// discard newer UUIDs
	uuids := make([]string, 0, len(fa.uuids))
	for _, uuid := range fa.uuids {
		_, suffix, _ := utils.ParseSuffixForUUID(uuid)
		if suffix > endSuffix {
			break
		}
		uuids = append(uuids, uuid)
	}
	fa.uuids = uuids
}

func (fa *filenameArgs) String() string {
	return fmt.Sprintf("(RelayBaseDir: %s, Filename: %s, SubDir: %s, UUIDs: %s, SafeRelayLog: %s)",
		fa.relayBaseDir, fa.filename, fa.subDir, strings.Join(fa.uuids, ";"), fa.safeRelayLog)
}

// filenameStrategy represents a relay purge strategy by filename
// similar to `PURGE BINARY LOGS TO`
type filenameStrategy struct {
	purging sync2.AtomicInt32
}

func newFilenameStrategy() PurgeStrategy {
	return &filenameStrategy{}
}

func (s *filenameStrategy) Check(args interface{}) (bool, error) {
	// do not support purge in the background
	return false, nil
}

func (s *filenameStrategy) Do(args interface{}) error {
	if !s.purging.CompareAndSwap(0, 1) {
		return ErrSelfPurging
	}
	defer s.purging.Set(0)

	fa, ok := args.(*filenameArgs)
	if !ok {
		return errors.NotValidf("args (%T) %+v", args, args)
	}

	return errors.Trace(purgeRelayFilesBeforeFile(fa.relayBaseDir, fa.uuids, fa.safeRelayLog))
}

func (s *filenameStrategy) Purging() bool {
	return s.purging.Get() > 0
}

func (s *filenameStrategy) Type() strategyType {
	return strategyFilename
}
