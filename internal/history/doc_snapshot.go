// Package history provides snapshot functionality for capturing and restoring
// point-in-time copies of the execution history store.
//
// # Snapshots
//
// A Snapshot is a complete, versioned copy of all history records serialised
// as JSON.  Snapshots can be used for:
//
//   - Backup before a destructive operation (e.g. migration, prune)
//   - Transferring history between machines
//   - Offline analysis without touching the live store
//
// # Usage
//
//	snap, err := history.TakeSnapshot(store, "/var/backups/cronwrap-snap.json")
//
//	loaded, err := history.LoadSnapshot("/var/backups/cronwrap-snap.json")
//
//	err = history.RestoreSnapshot(loaded, store)
package history
